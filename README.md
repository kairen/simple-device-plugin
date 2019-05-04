[![Build Status](https://travis-ci.org/kubedev/simple-device-plugin.svg?branch=master)](https://travis-ci.org/kubedev/simple-device-plugin) [![codecov](https://codecov.io/gh/kubedev/simple-device-plugin/branch/master/graph/badge.svg)](https://codecov.io/gh/kubedev/simple-device-plugin) [![Docker Pulls](https://img.shields.io/docker/pulls/kubedev/simple-device-plugin.svg)](https://hub.docker.com/r/kubedev/simple-device-plugin/)
# Simple Device Plugin
Learning how to implement a Kubernetes device-plugin. This device-plugin will automatically maps the SATA device according to your container SATA requirement.

## Prerequisites
The list of prerequisites for running the SATA device-plugin is described below:
* Kubernetes version = 1.10.x.
* The `DevicePlugins` feature gate enabled.

## Quick Start
To install the SATA device-plugin:
```sh
$ kubectl apply -f https://raw.githubusercontent.com/kubedev/device-plugin/master/artifacts/device-plugin.yml
$ kubectl -n kube-system get po -l name=device-plugin
NAME                            READY     STATUS    RESTARTS   AGE
device-plugin-ds-jlj8k   1/1       Running   0          38s
device-plugin-ds-sn2ff   1/1       Running   0          38s
```

To run the SATA pod:
```sh
$ kubectl apply -f https://raw.githubusercontent.com/kubedev/device-plugin/master/artifacts/test-device-pod.yml
$ kubectl get po
NAME              READY     STATUS    RESTARTS   AGE
test-device-pod   1/1       Running   0          30s

$ kubectl exec -ti test-device-pod sh
/ # ls /dev/ | grep "sd[a-z]"
sdb
/ # mkfs.vfat /dev/sdb
/ # od -vAn -N4 -tu4 < /dev/sdb
 1838176491
```
