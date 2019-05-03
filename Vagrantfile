Vagrant.require_version ">= 1.7.0"

Vagrant.configure("2") do |config|
  config.vm.provider "virtualbox"
  config.vm.box = "bento/ubuntu-16.04"

  private_count = 11
  (1..3).each do |mid|
    name = (mid <= 1) ? "k8s-m" : "k8s-n"
    id = (mid <= 1) ? mid : mid - 1
    config.vm.define "#{name}#{id}" do |n|
      n.vm.hostname = "#{name}#{id}"
      ip_addr = "192.16.35.#{private_count}"
      parityDisk = "#{n.vm.hostname}-disk"
      n.vm.network :private_network, ip: "#{ip_addr}", auto_config: true
      n.vm.provider :virtualbox do |vb, override|
        vb.name = "#{n.vm.hostname}"
        vb.gui = false
        vb.memory = 2048
        vb.cpus = 1
        if not File.exists?(parityDisk)
          vb.customize ['createhd', '--filename', parityDisk, '--size', 5 * 1024]
        end
        vb.customize ['storageattach', :id, '--storagectl', 'SATA Controller', '--port', 1, '--device', 0, '--type', 'hdd', '--medium', "#{parityDisk}.vdi"]
      end
      private_count += 1
    end
  end
  config.vm.provision :shell, path: "./hack/setup-vm.sh"
end
