package main

import "fmt"

func initializeXhyve(verbose bool) {
	//Download block0.img -> ~/.pot/xhyve/block0.img
	//download userboot.so -> ~/.pot/xhyve/userboot.so
	//Check if runfile exists
	//Create run file
	runFile := `
	#/bin/sh
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

	fmt.Println(runFile)
}
