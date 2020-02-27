package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/bmatcuk/go-vagrant"
)

//TODO: May need to rename this functions changing Vagrant to VM now that it includes xhyve.

func initializeVagrant(vmType string, ip string, verbose bool) {

	vagrantDirPath := getVagrantDirPath()
	vagrantFilePath := vagrantDirPath + "/Vagrantfile"

	_, err := os.Stat(vagrantFilePath)
	if os.IsNotExist(err) {
		fmt.Println("==> VagrantFile did not exists.\n==> Generating a new one on ~/.pot/Vagrantfile\n==> To change the Vagrant properties edit this file.\n==> After editing run: pot machine reload")
		var defaultVagrant string
		if vmType == "virtualbox" {
			if config.Memory == "" {
				config.Memory = "1024"
			}

			if config.Cpus == "" {
				config.Cpus = "1"
			}

			defaultVagrant = `Vagrant.configure("2") do |config|
  config.vm.box = "ebarriosjr/FreeBSD12.1-zfs"
  config.vm.box_version = "0.0.1"
  config.vm.network "private_network", ip: "` + ip + `"
  config.vm.define :potMachine do |t|
  end
  config.vm.provider "virtualbox" do |vb|
	vb.memory = "` + config.Memory + `"
	vb.cpus = ` + config.Cpus + `
  end
  config.vm.synced_folder ".", "/vagrant", disabled: true 
  config.vm.synced_folder "` + vagrantDirPath + `", "/vagrant", create: true, type: "nfs"
end`
			fmt.Println("==> Virtualbox ip: ", ip)

		} else if vmType == "xhyve" {
			if config.Memory == "" {
				config.Memory = "1024"
			}

			if config.Cpus == "" {
				config.Cpus = "1"
			}
			defaultVagrant = `
$script = <<-SCRIPT
echo 'ifconfig_em1="inet ` + ip + ` netmask 255.255.255.0"' >> /etc/rc.conf
SCRIPT

Vagrant.configure("2") do |config|
  config.vm.box = "ebarriosjr/FreeBSD12.1-zfs"
  config.vm.box_version = "0.0.1"
  config.vm.network "private_network", ip: "` + ip + `"
  config.vm.provision "shell", inline: $script
  config.vm.define :potMachine do |t|
  end
  config.vm.provider "xhyve" do |xhyve|
  	xhyve.cpus = ` + config.Cpus + `
  	xhyve.memory = "` + config.Memory + `"
  	xhyve.kernel_command=""
  end
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder "` + vagrantDirPath + `", "/vagrant", create: true, type: "nfs" 
end`
		} else if vmType == "nomad" {
			if config.Memory == "" {
				config.Memory = "1024"
			}

			if config.Cpus == "" {
				config.Cpus = "1"
			}

			defaultVagrant = `
$script = <<-SCRIPT
minipot-init -i ` + ip + ` && minipot-start
echo "POT_EXTRA_EXTIF=em1" >> /usr/local/etc/pot/pot.conf
echo "POT_NETWORK_em1=` + ip + `/24" >> /usr/local/etc/pot/pot.conf
chmod +x /usr/local/etc/pot/flavours/*
SCRIPT

Vagrant.configure("2") do |config|
  config.vm.box = "ebarriosjr/FreeBSD12-miniPot"
  config.vm.box_version = "0.0.2"
  config.vm.network "private_network", ip: "` + ip + `"
  config.vm.provision "shell", inline: $script
  config.vm.define :potMachine do |t|
  end
  config.vm.provider "virtualbox" do |vb|
	vb.memory = "` + config.Memory + `"
	vb.cpus = ` + config.Cpus + `
  end
  config.vm.synced_folder ".", "/vagrant", disabled: true 
  config.vm.synced_folder "` + vagrantDirPath + `", "/vagrant", create: true, type: "nfs"
end`
		} else {
			defaultVagrant = `Vagrant.configure("2") do |config|
  config.vm.box = "ebarriosjr/FreeBSD12.1-zfs"
  config.vm.box_version = "0.0.1"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder "` + vagrantDirPath + `", "/vagrant", create: true
  config.vm.define :potMachine do |t|
  end
  config.vm.provider "libvirt" do |lv|
     lv.memory = "` + config.Memory + `"
     lv.cpus = ` + config.Cpus + `
  end
end`
		}
		//Write Vagrantfile
		err = ioutil.WriteFile(vagrantFilePath, []byte(defaultVagrant), 0664)
		if err != nil {
			fmt.Println("ERROR: Error writting file to disk with err: \n", err)
			return
		}
	}

	client, err := vagrant.NewVagrantClient(vagrantDirPath)
	if err != nil {
		fmt.Println("ERROR: Error: No enviroment detected.")
		return
	}

	upCmd := client.Up()
	upCmd.Verbose = verbose
	if err := upCmd.Run(); err != nil {
		fmt.Println("ERROR: Error running vagrant with err: ", err)
	}
	if upCmd.Error != nil {
		fmt.Println("ERROR: Upcmd error: ", err)
	}

	if vmType == "nomad" {
		nomadAddr := "NOMAD_ADDR=http://" + ip + ":4646"
		fmt.Println("")
		fmt.Println("==> Pot machine created successfully. ")
		fmt.Println("==> ENV variables needed for Nomad: ")
		fmt.Println("==> export ", nomadAddr)
		fmt.Println("==> Extra info: ")
		fmt.Println("==> Region: global")
		fmt.Println("==> Datacenter: minipot")
		fmt.Println("")
	}
}

func redirectToVagrant(args []string) {
	vagrantDirPath := getVagrantDirPath()
	configSSHFile := vagrantDirPath + "sshConfig"
	if _, err := os.Stat(configSSHFile); os.IsNotExist(err) {
		createConfigCmd := "cd " + vagrantDirPath + "&& vagrant ssh-config > " + configSSHFile
		configCmd := exec.Command("bash", "-c", createConfigCmd)
		configCmd.Stdout = os.Stdout
		configCmd.Stdin = os.Stdin
		configCmd.Stderr = os.Stderr
		configCmd.Run()
		configCmd.Wait()
	}

	arguments := strings.Join(args, " ")
	termCmd := "ssh -F " + configSSHFile + " potMachine sudo " + arguments

	cmd := exec.Command("bash", "-c", termCmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	cmd.Wait()
}

func connectToVagrant() {
	vagrantDirPath := getVagrantDirPath()
	configSSHFile := vagrantDirPath + "sshConfig"
	VMType := checkVMType()
	if VMType == "vagrant" {
		if _, err := os.Stat(configSSHFile); os.IsNotExist(err) {
			createConfigCmd := "cd " + vagrantDirPath + "&& vagrant ssh-config > " + configSSHFile
			configCmd := exec.Command("bash", "-c", createConfigCmd)
			configCmd.Stdout = os.Stdout
			configCmd.Stdin = os.Stdin
			configCmd.Stderr = os.Stderr
			configCmd.Run()
			configCmd.Wait()
		}
	} else if VMType == "xhyve" {
		if _, err := os.Stat(configSSHFile); os.IsNotExist(err) {
			log.Fatal("ERROR: sshConfig file does not exists for xhyve machine")
		}
	}

	termCmd := "ssh -F " + configSSHFile + " potMachine "
	cmd := exec.Command("bash", "-c", termCmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	cmd.Wait()
}

func checkVMType() string {
	termCmd := "ps aux | grep potMachine"
	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error getting information on the VM type with err: ", err)
		log.Fatal(err)
	}
	cmd.Wait()
	commandResult := out.String()

	//ps aux | grep potMachine | grep virtualbox
	if strings.Contains(commandResult, "xhyve") {
		return "xhyve"
	}

	return "vagrant"
}

func destroyVagrant(verbose bool) {
	vagrantDirPath := getVagrantDirPath()

	VMType := checkVMType()
	if VMType == "vagrant" {

		client, err := vagrant.NewVagrantClient(vagrantDirPath)
		if err != nil {
			fmt.Println("ERROR: Error: No enviroment detected.")
			return
		}
		destroyCmd := client.Destroy()
		destroyCmd.Verbose = verbose
		if err := destroyCmd.Run(); err != nil {
			fmt.Println("ERROR: Error running vagrant with err: ", err)
			return
		}
		if destroyCmd.Error != nil {
			fmt.Println("ERROR: Upcmd error: ", err)
			return
		}

		err = os.Remove(vagrantDirPath + "Vagrantfile")
		if err != nil {
			fmt.Println("Error removing Vagrantfile from ", vagrantDirPath)
		}

		os.Remove(vagrantDirPath + "sshConfig")
	} else if VMType == "xhyve" {
		//Connect to xhyve and poweroff
		redirectToVagrant([]string{"poweroff"})

		//TODO: Remove all files from ~/.pot/xhyve
	}
}

func getVagrantDirPath() string {
	homeDir := getUserHome()
	if homeDir == "" {
		fmt.Println("ERROR: Error getting user home directory info...")
		return ""
	}

	vagrantDirPath := homeDir + "/.pot/"
	if _, err := os.Stat(vagrantDirPath); os.IsNotExist(err) {
		os.Mkdir(vagrantDirPath, 0755)
	}

	return vagrantDirPath
}

func getUserHome() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user info with err: \n", err)
		return ""
	}
	return usr.HomeDir
}

func startVagrant(verbose bool) {
	vagrantDirPath := getVagrantDirPath()
	//TODO: Check if vm is not vagrant
	if _, err := os.Stat(vagrantDirPath + "/xhyve/runFreeBSD.sh"); os.IsNotExist(err) {

		client, err := vagrant.NewVagrantClient(vagrantDirPath)
		if err != nil {
			fmt.Println("ERROR: Error starting potMachine")
			return
		}

		upCmd := client.Up()
		upCmd.Verbose = verbose
		if err := upCmd.Run(); err != nil {
			fmt.Println("ERROR: Error running vagrant with err: ", err)
		}
		if upCmd.Error != nil {
			fmt.Println("ERROR: Startcmd error: ", err)
		}
	} else {
		err := runXhyve()
		if err != nil {
			log.Fatal("ERROR: Can not start xhyve VM with err: ", err)
		}
	}
}

func stopVagrant(verbose bool) {
	vagrantDirPath := getVagrantDirPath()

	VMType := checkVMType()
	if VMType == "vagrant" {

		client, err := vagrant.NewVagrantClient(vagrantDirPath)
		if err != nil {
			fmt.Println("ERROR: Error stoping potMachine")
			return
		}

		upCmd := client.Halt()
		upCmd.Verbose = verbose
		if err := upCmd.Run(); err != nil {
			fmt.Println("ERROR: Error running vagrant with err: ", err)
		}
		if upCmd.Error != nil {
			fmt.Println("ERROR: Stopcmd error: ", err)
		}
	} else if VMType == "xhyve" {
		//Connect to xhyve and poweroff
		redirectToVagrant([]string{"poweroff"})
	}
}

func reloadVagrant(verbose bool) {
	vagrantDirPath := getVagrantDirPath()
	//TODO: Check if vm is not vagrant
	VMType := checkVMType()
	if VMType == "vagrant" {
		client, err := vagrant.NewVagrantClient(vagrantDirPath)
		if err != nil {
			fmt.Println("ERROR: Error reloading potMachine")
			return
		}

		upCmd := client.Reload()
		upCmd.Verbose = verbose
		if err := upCmd.Run(); err != nil {
			fmt.Println("ERROR: Error running vagrant with err: ", err)
		}
		if upCmd.Error != nil {
			fmt.Println("ERROR: Stopcmd error: ", err)
		}
	} else if VMType == "xhyve" {
		//Stop xhyve VM
		redirectToVagrant([]string{"poweroff"})
		//Start xhyve VM
		err := runXhyve()
		if err != nil {
			log.Fatal("ERROR: Xhyve is not able to start with err: ", err)
		}
	}
}
