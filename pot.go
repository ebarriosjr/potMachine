package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func redirectToPot(args []string) {
	vagrantDirPath := getVagrantDirPath()
	configSSHFile := vagrantDirPath + "sshConfig"
	if _, err := os.Stat(configSSHFile); os.IsNotExist(err) {
		createConfigCmd := "cd " + vagrantDirPath + "&& vagrant ssh-config > " + configSSHFile
		configCmd := exec.Command("bash", "-c", createConfigCmd)

		if err := configCmd.Start(); err != nil {
			fmt.Println("cmd.Start: ", err)
			return
		}

		if err := configCmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0

				if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					fmt.Println("==> No enviroment set on this machine or Xhyve VM not running\n==> Please run pot machine init [virtualbox/libvirt/xhyve/nomad]\n==> Or \"pot machine start\" if xhyve is initialized on the machine")
					err = os.Remove(vagrantDirPath + "sshConfig")
					if err != nil {
						fmt.Println("Error removing ssConfig file from ", vagrantDirPath)
					}
					os.Exit(1)
				}
			} else {
				fmt.Println("cmd.Wait: ", err)
				return
			}
		}

	}

	arguments := strings.Join(args, " ")
	//fmt.Println("Arguments: ", arguments)

	termCmd := "ssh -t -q -F " + configSSHFile + " potMachine sudo pot " + arguments

	cmd := exec.Command("bash", "-c", termCmd)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	cmd.Wait()

	if err != nil {
		if err.Error() == "255" {
			fmt.Println("==> Not able to connect to the potMachine.")
			fmt.Println("==> Maybe the potMachine is not runnig.")
			fmt.Println("==> Try: \"pot machine start\" or \"pot machine init -h\" ")
		}
	}
}

func createFlavour(flavour string) {
	basePath := getVagrantDirPath()
	flavourPath := basePath + flavour

	cmd := exec.Command(config.Editor, flavourPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running Editor with err: ", err)
	}

	flavourVagrantPath := "/vagrant/" + flavour

	redirectToVagrant([]string{"cp", flavourVagrantPath, "/usr/local/etc/pot/flavours/"})

	flavourPotPath := "/usr/local/etc/pot/flavours/" + flavour

	redirectToVagrant([]string{"chmod", "755", flavourPotPath})

}
