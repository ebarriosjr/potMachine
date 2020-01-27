package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	randomdata "github.com/Pallinder/go-randomdata"
	log "github.com/Sirupsen/logrus"
)

func buildPot(tag *string) {
	vagrantDirPath := getVagrantDirPath()
	potFile, err := os.Open("./Potfile")
	if err != nil {
		log.Error("No Potfile on path")
		return
	}
	defer potFile.Close()

	potMap := make(map[string][]string)
	var copySlice []copy
	var runSlice []string
	var envSlice []env
	var argSlice []env
	var addSlice []string
	var baseCommand string

	scanner := bufio.NewScanner(potFile)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") && (strings.Contains(line, "ENV") || strings.Contains(line, "ARG")) {
			line = strings.Replace(line, "=", " ", -1)
		}

		split := strings.Split(line, " ")
		switch option := split[0]; option {
		case "RUN":
			runSlice = append(runSlice, strings.Join(split[1:], " "))
		case "COPY":
			copySlice = append(copySlice, copy{source: split[1], destination: split[2]})
		case "ENV":
			envSlice = append(envSlice, env{variable: split[1], value: split[2]})
		case "ARG":
			argSlice = append(argSlice, env{variable: split[1], value: split[2]})
		case "ADD":
			addSlice = append(addSlice, split[1])
		case "CMD":
			baseCommand = strings.Join(split[1:], "")
		default:
			potMap[split[0]] = split[1:]
		}
	}

	if len(potMap["FROM"]) == 0 {
		log.Error("FROM can't be empty...")
		return
	}

	if len(baseCommand) == 0 {
		log.Error("CMD can't be emtpy...")
		return
	}

	if len(potMap["NAME"]) == 0 {
		log.Info("==> NAME is empty, generating a random name...")
		randomName := randomdata.SillyName()
		potMap["NAME"] = []string{randomName}
		log.Info("NAME: ", potMap["NAME"][0])
	}

	//Initialize run.sh file
	var runFile string
	var fileContent string
	if len(runSlice) > 0 || len(argSlice) > 0 {
		runFile = vagrantDirPath + "/run.sh"
		fileContent = "#!/bin/sh\n"
	}

	//Add ARGS to the run.sh
	if len(argSlice) > 0 {
		for _, env := range argSlice {
			fileContent = fileContent + "export " + env.variable + "=\"" + env.value + "\"\n"
		}
	}

	if len(addSlice) > 0 {
		for i := range addSlice {
			fileContent = fileContent + "fetch \"" + addSlice[i] + "\"\n"
		}
	}

	//Add RUN commands to the flavour
	if len(runSlice) > 0 {
		for i := range runSlice {
			fileContent = fileContent + runSlice[i] + "\n"
		}
	}

	if len(runSlice) > 0 || len(argSlice) > 0 {
		newFlavorByte := []byte(fileContent)
		err := ioutil.WriteFile(runFile, newFlavorByte, 0755)
		if err != nil {
			log.Error("Error writting file to disk!")
			return
		}

		flavorCommand := " cp /vagrant/run.sh /usr/local/etc/pot/flavours/run.sh"
		redirectToVagrant([]string{flavorCommand})

		//delete flavour file...

		err = os.Remove(runFile)

		if err != nil {
			log.Error("Error removing file run.sh with err: ", err)
			return
		}
	}

	//Create pot
	var create string
	if len(runSlice) > 0 {
		if len(potMap["EXPOSE"]) > 0 {
			create = "create -p " + potMap["NAME"][0] + " -b " + potMap["FROM"][0] + " -t single -N public-bridge -f run"
		} else {
			create = "create -p " + potMap["NAME"][0] + " -b " + potMap["FROM"][0] + " -t single -f run"
		}
	} else {
		create = "create -p " + potMap["NAME"][0] + " -b " + potMap["FROM"][0] + " -t single"
	}

	//Adding extra flavours in the orther they appear on the Potfile.
	if len(potMap["FLAVOUR"]) > 0 {
		for _, flavour := range potMap["FLAVOUR"] {
			create = create + " -f " + flavour
		}
	}

	//Create pot with flavours
	redirectToPot([]string{create})

	//delete flavour
	deleteRunFlavour := "sudo rm -rf /usr/local/etc/pot/flavours/run.sh"
	redirectToVagrant([]string{deleteRunFlavour})

	if len(copySlice) > 0 {
		for i := range copySlice {
			copyDirPath := vagrantDirPath + "copy"
			// mkdir .pot/copy --- os.Mkdir()
			err := os.Mkdir(copyDirPath, 0775)
			if err != nil {
				log.Error("Error creating intermediate directory for copy command...")
				return
			}

			var source string
			if copySlice[i].source == "." {
				source, _ = os.Getwd()
				source = source + "/*"
			} else if strings.HasPrefix(copySlice[i].source, "/") {
				source = copySlice[i].source
			} else {
				pwd, _ := os.Getwd()
				source = pwd + "/" + copySlice[i].source
			}

			copyCmd := "cp -r " + source + " " + copyDirPath
			_, err = exec.Command("bash", "-c", copyCmd).Output()
			if err != nil {
				log.Error("Error copying files...")
			}

			var origin string
			if copySlice[i].source == "." {
				origin = "/vagrant/copy/"
			} else {
				origin = "/vagrant/copy/" + copySlice[i].source
			}
			destination := copySlice[i].destination
			copyCommand := "copy-in -p " + potMap["NAME"][0] + " -s " + origin + " -d " + destination
			redirectToPot([]string{copyCommand})

			os.RemoveAll(copyDirPath)
		}
	}

	if len(envSlice) > 0 {
		setEnv := "set-env -p " + potMap["NAME"][0]
		for _, env := range envSlice {
			setEnv = setEnv + " -E " + env.variable + "=" + env.value
		}
		redirectToPot([]string{setEnv})
	}

	//NEEDS WORK
	if len(potMap["EXPOSE"]) > 0 {
		setPort := "export-ports -p " + potMap["NAME"][0] + " -e " + potMap["EXPOSE"][0]
		redirectToPot([]string{setPort})
	}

	if len(potMap["CPU"]) > 0 {
		setCPU := "set-rss -p " + potMap["NAME"][0] + " -C " + potMap["CPU"][0]
		redirectToPot([]string{setCPU})
	}

	if len(potMap["MEMORY"]) > 0 {
		setMem := "set-rss -p " + potMap["NAME"][0] + " -M " + potMap["MEMORY"][0]
		redirectToPot([]string{setMem})
	}

	// Pot set CMD
	// Needs to be imported as an array. Stop beeing lazy!
	if len(potMap["CMD"]) > 0 {
		setNoRcScript := "set-attr -p " + potMap["NAME"][0] + " -A no-rc-script -V on"
		redirectToPot([]string{setNoRcScript})

		var commandSlice []string
		err := json.Unmarshal([]byte(baseCommand), &commandSlice)
		if err != nil {
			fmt.Println("Unable to convert CMD to slice with err: ", err)
			return
		}

		command := strings.Join(commandSlice, " ")
		setPotCmd := "set-cmd -p " + potMap["NAME"][0] + " -c " + command
		redirectToPot([]string{setPotCmd})
	}
	//redirectToPot([]string{startCmd})

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if *tag != "" {

		// artifactory.tcs.trv.cloud:9090/project/name:0.1
		splitVersion := strings.Split(*tag, ":")
		lengthVersion := len(splitVersion) - 1
		version := splitVersion[lengthVersion]

		// Create exports folder if it does not exists
		mkdirCommand := "mkdir -p /vagrant/exports"
		redirectToVagrant([]string{mkdirCommand})

		// Export pot to /vagrant/exports folder (NFS Mount)
		exportCommand := "export -p " + potMap["NAME"][0] + " -t " + version + " -D /vagrant/exports -A"
		redirectToPot([]string{exportCommand})
	}

	/*
		('add', 'arg', 'cmd', 'copy', 'entrypoint', 'env', 'expose', 'from', 'healthcheck',
		 'label', 'maintainer', 'onbuild', 'run', 'shell', 'stopsignal', 'user', 'volume', 'workdir')
	*/

}
