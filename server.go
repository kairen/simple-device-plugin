package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	resourceName  = "kubedev.org/sata"
	serverSock    = pluginapi.DevicePluginPath + "sata.sock"
	numberDevices = 5
)

// DevicePlugin implements the Kubernetes device plugin API
type DevicePlugin struct {
	devs   []*pluginapi.Device
	socket string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewDevicePlugin returns an initialized DevicePlugin
func NewDevicePlugin() (*DevicePlugin, error) {
	devices, err := GetDevices()
	if err != nil {
		return nil, err
	}

	var devs = make([]*pluginapi.Device, len(devices))
	for i := range devs {
		devs[i] = &pluginapi.Device{
			ID:     fmt.Sprint(i),
			Health: pluginapi.Healthy,
		}
	}

	return &DevicePlugin{
		devs:   devs,
		socket: serverSock,
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}, nil
}

func (m *DevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *DevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

	// Wait for server to start by launching a blocking connection
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	// go m.healthcheck()

	return nil
}

// Stop stops the gRPC server
func (m *DevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}
	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *DevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}

	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (m *DevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	glog.Infof("Exposing devices: %v", m.devs)
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			// FIXME: there is no way to recover from the Unhealthy state.
			d.Health = pluginapi.Unhealthy
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
		}
	}
}

// Allocate which return list of devices.
func (m *DevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	glog.Infof("Allocate request: %v", reqs)

	devices, _ := GetDevices()
	responses := pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		ds := make([]*pluginapi.DeviceSpec, len(req.DevicesIDs))
		response := pluginapi.ContainerAllocateResponse{Devices: ds}

		for i := range req.DevicesIDs {
			ds[i] = &pluginapi.DeviceSpec{
				HostPath:      devices[i].Path,
				ContainerPath: devices[i].Path,
				Permissions:   "rwm",
			}
		}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}
	glog.Infof("Allocate response: %v", responses)
	return &responses, nil
}

func (m *DevicePlugin) unhealthy(dev *pluginapi.Device) {
	m.health <- dev
}

func (m *DevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (m *DevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// func (m *DevicePlugin) healthcheck() {
//
// }

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *DevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		glog.Errorf("Could not start device plugin: %v", err)
		return err
	}
	glog.Infof("Starting to serve on %s", m.socket)

	err = m.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		glog.Errorf("Could not register device plugin: %v", err)
		m.Stop()
		return err
	}
	glog.Infof("Registered device plugin with Kubelet")
	return nil
}
