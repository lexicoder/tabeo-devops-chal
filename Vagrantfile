# -*- mode: ruby -*-
# vi: set ft=ruby :
default_box = "generic/opensuse42"

Vagrant.configure("2") do |config|
  config.vm.define "k8" do |k8|
    k8.vm.box = default_box
    k8.vm.hostname = "master"
    k8.vm.network 'private_network', ip: "192.168.0.200",  virtualbox__intnet: true
    k8.vm.network "forwarded_port", guest: 22, host: 2222, id: "ssh", disabled: true
    k8.vm.network "forwarded_port", guest: 22, host: 2000 # Master Node SSH
    k8.vm.network "forwarded_port", guest: 6443, host: 6443 # API Access
    k8.vm.provider "virtualbox" do |v|
      v.memory = "2048"
      v.name = "k8"
      end
    k8.vm.provision "file", source: "./get-k3s.sh", destination: "./get-k3s.sh"
    k8.vm.provision "shell", inline: <<-SHELL
      sudo update-ca-certificates
      sudo zypper refresh
      sudo zypper --non-interactive install bzip2
      sudo zypper --non-interactive install etcd
      sudo zypper --non-interactive install apparmor-parser
      sh /home/vagrant/get-k3s.sh
    SHELL
  end
end
