# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure(2) do |config|
    config.vm.box = "ubuntu/xenial64"
    config.vm.box_check_update = false
    config.hostmanager.enabled = true
    config.hostmanager.manage_host = true
    config.hostmanager.ignore_private_ip = false
    config.hostmanager.include_offline = false

    config.vm.define :juno_test do |juno_test|
        juno_test.vm.host_name = "local-juno-test"
        juno_test.hostmanager.aliases = ["local.juno-test"]
    end

    # Provider-specific configuration so you can fine-tune various
    # backing providers for Vagrant. These expose provider-specific options.
    # Example for VirtualBox:
    #
    config.vm.provider "virtualbox" do |vb|
    # Display the VirtualBox GUI when booting the machine
    vb.gui = false

    # Customize the amount of memory on the VM:
    vb.memory = "1024"
    end
    config.vm.provision "shell", inline: <<-SHELL
        set -xeuo pipefail
        sysctl net.ipv6.conf.all.forwarding=1
        apt-get update
        apt-get install -y apt-transport-https software-properties-common aptitude
        # clickhouse
        apt-key adv --keyserver keyserver.ubuntu.com --recv-keys E0C56BD4
        # gophers
        apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 136221EE520DDFAF0A905689B9316A7BC7917B12
        # redis
        apt-key adv --keyserver keyserver.ubuntu.com --recv-keys C73998DC9DFEA6DCF1241057308C15A29AD198E9
        # docker
        apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 8D81803C0EBFCD88
        add-apt-repository "deb https://download.docker.com/linux/ubuntu xenial edge"
        add-apt-repository ppa:chris-lea/redis-server
        add-apt-repository ppa:gophers/archive
        apt-get update
        apt-get install -y golang-1.8
        apt-get install -y docker-ce
        apt-get install -y htop ethtool mc
        apt-get install -y python-pip
        apt-get install -y redis-tools
        pip install -U docker-compose
        ln -nvsf /usr/lib/go-1.8/bin/go /bin/go
        ln -nvsf /usr/lib/go-1.8/bin/gofmt /bin/gofmt
        bash -c "cd /vagrant/src && GOPATH=/vagrant go get -v ./..."
        set +x
        echo "juno-test PROVISIONING DONE, use folloding scenario for developing"
        echo "#  vagrant ssh juno-test"
        echo "#  bash -c \\\"cd /vagrant && GOPATH=/vagrant go run cmd/kv_server.go \\\""
        echo "for docker build run following command"
        echo "#  cd /vagrant && sudo ./run_docker.sh"
        echo "Good Luck ;)"
    SHELL
end
