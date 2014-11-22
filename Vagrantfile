# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.network :forwarded_port, guest: 3000, host: 3000

  config.vm.provider "virtualbox" do |vb|
    # vb.gui = true
    vb.customize ["modifyvm", :id, "--memory", "512"]
  end

  config.vm.provider :lxc do |lxc, override|
    override.vm.box = "trusty64-lxc"
    override.vm.box_url = "https://vagrantcloud.com/fgrehm/boxes/trusty64-lxc/versions/2/providers/lxc.box"
    lxc.backingstore = "btrfs"
  end

  config.vm.provision :ansible do |ansible|
    ansible.playbook = "playbooks/playbook.yml"
    ansible.sudo = true
    ansible.extra_vars = { ansible_ssh_user: 'vagrant' }
  end
end
