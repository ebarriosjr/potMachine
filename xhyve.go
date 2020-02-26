package main

import "fmt"

func initializeXhyve(verbose bool) {
	//mkdir ~/.pot/xhyve if it doesnt exists
	//Download from github respo xhyve.tar.zg -> ~/.pot/xhyve
	//untar xhyve.tar.gz into ~/.pot/xhyve
	//Check if runfile exists
	//Create run file
	runFile := `#/bin/sh
UUID="-U potpotpo-potp-potp-potp-potmachinepp"
USERBOOT="~/.pot/xhyve/userboot.so"
IMG="~/.pot/xhyve/block0.img"
KERNELENV=""

MEM="-m 4G"
SMP="-c 2"
PCI_DEV="-s 0:0,hostbridge -s 31,lpc"
NET="-s 2:0.virtio-net"
IMG_HDD="-s 4:0,virtio-blk,$IMG"
LPC_DEV="-l com1,stdio"
ACPI="-A"

xhyve $ACPI $MEM $SMP $PCI_DEV $LPC_DEV $NET $IMG_HDD $UUID -f fbsd,$USERBOOT,$IMG,"$KERNELENV"
`
	//Write runfile to ~/.pot/xhyve/run.sh
	//Initializa xhyve vm
	//Get xhyve ip
	xhyveIP := ""
	//generate sshConfig file
	sshConfig := `Host potMachine
  HostName ` + xhyveIP + `
  User vagrant
  Port 22
  UserKnownHostsFile /dev/null
  StrictHostKeyChecking no
  PasswordAuthentication no
  IdentityFile ~/.pot/xhyve/private_key
  IdentitiesOnly yes
  LogLevel FATAL
`

	fmt.Println("Runfile", runFile)
	fmt.Println("sshConfig", sshConfig)
}
