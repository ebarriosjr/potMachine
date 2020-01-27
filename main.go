package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
)

var config Config

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "which "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func init() {
	// Read config File and parse it.
	configPath := getVagrantDirPath()
	configFile := configPath + "potConfig"
	_, err := ioutil.ReadFile(configFile)
	if err != nil {

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		var vmType string
		if isCommandAvailable("virtualbox") {
			vmType = "virtualbox"
		} else if isCommandAvailable("libvirtd") {
			vmType = "libvirtd"
		}

		if vmType == "" {
			fmt.Println("==> Virtualbox or Libvirtd needs to be installed for potMachine to run.")
			os.Exit(1)
		}

		fmt.Println("==> No configuration file found.")
		fmt.Println("==> Creating a default one on:", configFile)
		fmt.Println("==> with editor:", editor)
		fmt.Println("==> with VM manager:", vmType)
		fmt.Println("==> with default ip: 192.168.44.100")
		fmt.Println("==> with default memory: 2048 MB")
		fmt.Println("==> with default cpus: 2")

		defaultConfig := `Editor = "` + editor + `"
VMType = "` + vmType + `"
IP = "192.168.44.100"
Memory = "2048"
Cpus = "2"
`
		//Write configFile
		err = ioutil.WriteFile(configFile, []byte(defaultConfig), 0664)
		if err != nil {
			fmt.Println("Error: Error writting file to disk with err: \n", err)
			return
		}

	}
	tomlData, _ := ioutil.ReadFile(configFile)

	if _, err := toml.Decode(string(tomlData), &config); err != nil {
		fmt.Println("Error running toml.Decode: ", err)
		return
	}

}

func main() {
	buildCommand := flag.NewFlagSet("build", flag.ExitOnError)
	buildCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot build [OPTIONS]\n\nBuild an image from a Potfile\n\nOptions:\n  -t -- Tag for the build")
	}

	pushCommand := flag.NewFlagSet("push", flag.ExitOnError)
	pushCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot push [OPTIONS] NAME[:TAG]\n\nPush an image to a remote host\n\nOptions:\n  -t -- Tag for the build")
	}

	loginCommand := flag.NewFlagSet("login", flag.ExitOnError)
	loginCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot login [Server] [OPTIONS]\n\nOptions:\n  -p              -- Password\n  -u              -- Username\n  -password-stdin -- Input password from stdin\n")
	}

	machineCommand := flag.NewFlagSet("machine", flag.ExitOnError)
	machineCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine [command]\n\nCommands:\n  init -- Initialize enviroment for pot with Vagrant\n  start -- Start the potMachine\n  stop -- Stop the potMachine\n  reload -- Reloads the potMachine\n  ssh -- Connects to the potMachine\n  destroy -- Removes the enviroment from the machine\n  add-flavour -- creates a pot flavor for the potMachine\n")
	}

	startCommand := flag.NewFlagSet("start", flag.ExitOnError)
	startCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine start [OPTIONS]\n\nStarts the local potMachine\n\nOptions:\n  -v -- Verbose\n")
	}

	stopCommand := flag.NewFlagSet("stop", flag.ExitOnError)
	stopCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine stop [OPTIONS]\n\nStops the local potMachine\n\nOptions:\n  -v -- Verbose\n")
	}

	reloadCommand := flag.NewFlagSet("reload", flag.ExitOnError)
	reloadCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine reload [OPTIONS]\n\nReloads the local potMachine\n\nOptions:\n  -v -- Verbose\n")
	}

	destroyCommand := flag.NewFlagSet("destroy", flag.ExitOnError)
	destroyCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine destroy [OPTIONS]\n\nDestroys the local potMachine\n\nOptions:\n  -v -- Verbose\n")
	}

	addFlavourCommand := flag.NewFlagSet("add-flavour", flag.ExitOnError)
	addFlavourCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine add-flavour [NAME]\n\nCreates a flavour inside the local potMachine\n")
	}

	initCommand := flag.NewFlagSet("init", flag.ExitOnError)
	initCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine init [VMTYPE] [OPTIONS]\n\nCreates a local potMachine\n\nVmType:\n  virtualbox -- Starts the potMachine in Virtualbox\n  libvirt -- Starts the potMachine inLibvirt\n\nOptions:\n  -ip -- Assigns an ip to the potMachine (only works with Virtualbox type)\n  -v -- Verbose\n")
	}

	sshCommand := flag.NewFlagSet("ssh", flag.ExitOnError)
	initCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine ssh \n\nConnects to the local potMachine\n\n")
	}

	initVBCommand := flag.NewFlagSet("virtualbox", flag.ExitOnError)
	initVBCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine init virtualbox [OPTIONS]\n\nCreates a local potMachine in Virtualbox\n\nOptions:\n  -ip -- Assigns an ip to the potMachine\n  -v -- Verbose\n")
	}

	initLibvirtCommand := flag.NewFlagSet("libvirt", flag.ExitOnError)
	initLibvirtCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine init libvirt [OPTIONS]\n\nCreates a local potMachine in Libvirt\n\nOptions:\n  -v -- Verbose\n")
	}

	initNomadCommand := flag.NewFlagSet("nomad", flag.ExitOnError)
	initNomadCommand.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: pot machine init nomad [OPTIONS]\n\nCreates a local potMachine in virtualbox with minipot installed\n\nOptions:\n  -ip -- Assigns an ip to the potMachine\n  -v -- Verbose\n")
	}

	tagBuildFlag := buildCommand.String("t", "", "Url, name and tag in the 'url/name:tag' format")
	tagPushFlag := pushCommand.String("t", "", "Url, name and tag in the 'url/name:tag' format")
	userFlag := loginCommand.String("u", "", "Username")
	passFlag := loginCommand.String("p", "", "Password")
	inlinePassFlag := loginCommand.Bool("password-stdin", false, "Take the password from stdin")
	VBFlag := initVBCommand.String("ip", "", "Ip address for the virtualbox VM")
	NomadFlag := initNomadCommand.String("ip", "", "Ip address for the nomad VM")
	verboseVB := initVBCommand.Bool("v", false, "Verbose")
	verboseLibvirt := initLibvirtCommand.Bool("v", false, "Verbose")
	verboseNomad := initNomadCommand.Bool("v", false, "Verbose")
	verboseDestroy := destroyCommand.Bool("v", false, "Verbose")
	verboseStart := startCommand.Bool("v", false, "Verbose")
	verboseStop := stopCommand.Bool("v", false, "Verbose")
	verboseReload := reloadCommand.Bool("v", false, "Verbose")

	if len(os.Args) == 1 {
		fmt.Println("Local Commands: \n\n    machine -- Commands to deal with the local potMachine env \n    build   -- Build an image from a Potfile\n    push    -- Push an Pot image to a web endpoint\n    login   -- Log in to a Pot file server\n ")
		fmt.Println("Remote Commands:\n ")
		redirectToPot(os.Args[1:])
		return
	}

	switch os.Args[1] {
	case "machine":
		machineCommand.Parse(os.Args[2:])
	case "build":
		buildCommand.Parse(os.Args[2:])
	case "push":
		pushCommand.Parse(os.Args[2:])
	case "login":
		loginCommand.Parse(os.Args[2:])
	default:
		if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "--h" {
			fmt.Println("Local Commands: \n\n    machine -- Commands to deal with the local potMachine env \n    build   -- Build an image from a Potfile\n    push    -- Push an Pot image to a web endpoint\n    login   -- Log in to a Pot file server\n ")
			fmt.Println("Remote Commands:\n ")
		}
		redirectToPot(os.Args[1:])
	}

	if buildCommand.Parsed() {
		fmt.Println("==> Building pot...")
		buildPot(tagBuildFlag)
		return
	}

	if pushCommand.Parsed() {
		if *tagPushFlag == "" {
			fmt.Println("Usage of push:\n  -t string\n        Url, name and tag in the 'url/name:tag' format")
			return
		}

		fmt.Println("==> Pushing pot to: ", *tagPushFlag)
		pushPot(tagPushFlag)
		return
	}

	if loginCommand.Parsed() {
		loginPot(loginCommand.Args(), userFlag, passFlag, inlinePassFlag)
		return
	}

	if machineCommand.Parsed() {
		if len(os.Args) < 3 || len(os.Args) == 0 {
			fmt.Println("Usage: pot machine [command]\n\nCommands:\n  init -- Initialize enviroment for pot with Vagrant\n  start -- Start the potMachine\n  stop -- Stop the potMachine\n  ssh -- Connects to the potMachine\n  destroy -- Removes the enviroment from the machine\n  add-flavour -- creates a pot flavor for the potMachine")
			return
		}
		switch os.Args[2] {
		case "destroy":
			destroyCommand.Parse(os.Args[3:])
		case "init":
			initCommand.Parse(os.Args[3:])
		case "ssh":
			sshCommand.Parse(os.Args[3:])
		case "start":
			startCommand.Parse(os.Args[3:])
		case "stop":
			stopCommand.Parse(os.Args[3:])
		case "reload":
			reloadCommand.Parse(os.Args[3:])
		case "add-flavour":
			addFlavourCommand.Parse(os.Args[3:])
		default:
			fmt.Println("Usage: pot machine [command]\n\nCommands:\n  init -- Initialize enviroment for pot with Vagrant\n  start -- Start the potMachine\n  stop -- Stop the potMachine\n  ssh -- Connects to the potMachine\n  destroy -- Removes the enviroment from the machine\n  add-flavour -- creates a pot flavor for the potMachine")
			return
		}
	}

	if destroyCommand.Parsed() {
		fmt.Println("==> Destroying potMachine")
		destroyVagrant(*verboseDestroy)
		return
	}

	if startCommand.Parsed() {
		fmt.Println("==> Starting potMachine")
		startVagrant(*verboseStart)
	}

	if stopCommand.Parsed() {
		fmt.Println("==> Stoping potMachine")
		stopVagrant(*verboseStop)
	}

	if reloadCommand.Parsed() {
		fmt.Println("==> Reloading potMachine")
		reloadVagrant(*verboseReload)
	}

	if addFlavourCommand.Parsed() {
		if len(os.Args) < 4 {
			fmt.Println("Need to attach a flavour name for creation.")
			return
		}
		fmt.Println("==> Creating new flavor: ", os.Args[3])
		createFlavour(os.Args[3])
	}

	if sshCommand.Parsed() {
		connectToVagrant()
	}

	if initCommand.Parsed() {
		if len(os.Args) < 4 || len(os.Args) == 0 {
			if config.VMType != "" {
				os.Args = append(os.Args, config.VMType)
			} else {
				fmt.Println("Usage: pot machine init [command]\n\nCommands:\n  virtualbox -- Initialize virtualbox enviroment for pot with Vagrant\n  libvirt -- Initialize libvirt enviroment for pot with Vagrant\n  nomad -- Initialize a virtualbox env with minipot running")
				return
			}
		}
		switch os.Args[3] {
		case "virtualbox":
			initVBCommand.Parse(os.Args[4:])
		case "libvirt":
			initLibvirtCommand.Parse(os.Args[4:])
		case "nomad":
			initNomadCommand.Parse(os.Args[4:])
		default:
			fmt.Println("Usage: pot machine init [command]\n\nCommands:\n  virtualbox -- Initialize virtualbox enviroment for pot with Vagrant\n  libvirt -- Initialize libvirt enviroment for pot with Vagrant\n  nomad -- Initialize a virtualbox env with minipot running")
			return
		}
	}

	if initVBCommand.Parsed() {
		fmt.Println("==> Initializing vagrant virtualbox potMachine for the first time....")
		if *VBFlag != "" {
			initializeVagrant("virtualbox", *VBFlag, *verboseVB)
		} else {
			if config.IP != "" {
				initializeVagrant("virtualbox", config.IP, *verboseVB)
			} else {
				fmt.Println("No ip passed as a flag or in the config file.")
				return
			}
		}
		return
	}

	if initLibvirtCommand.Parsed() {
		fmt.Println("==> Initializing vagrant libvirt potMachine for the first time....")
		initializeVagrant("libvirtd", "", *verboseLibvirt)
		return
	}

	if initNomadCommand.Parsed() {
		if *NomadFlag != "" {
			initializeVagrant("nomad", *NomadFlag, *verboseNomad)
		} else {
			if config.IP != "" {
				initializeVagrant("nomad", config.IP, *verboseNomad)
			} else {
				fmt.Println("No ip passed as a flag or in the config file.")
				return
			}
		}
		return
	}
}
