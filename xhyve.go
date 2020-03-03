package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/machinebox/progress"
	"github.com/pkg/errors"
)

var (
	xhyveIP string
)

func initializeXhyve(verbose bool) {
	potDirPath := getVagrantDirPath()
	xhyveDirPath := potDirPath + "/xhyve"

	if _, err := os.Stat(xhyveDirPath); os.IsNotExist(err) {
		fmt.Println("==> Creating ~/.pot/xhyve directory")
		os.Mkdir(xhyveDirPath, 0775)
	}

	//Download from github respo xhyve.tar.zg -> ~/.pot/xhyve
	fileURL := "https://app.vagrantup.com/ebarriosjr/boxes/FreeBSD12.1-zfs/versions/0.0.1/providers/xhyve.box"
	tarPath := xhyveDirPath + "/potMachine.tar.gz"

	fmt.Println("==> Checking if tar file already exists on ~/.pot/xhyve/potMachine.tar.gz")
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		fmt.Println("==> Downloading tar file to ~/.pot/xhyve/potMachine.tar.gz")
		if err := downloadFile(tarPath, fileURL); err != nil {
			fmt.Println("Error downloading tar file from vagrant cloud with err: ", err)
			log.Fatal()
		}
	}

	fmt.Println("==> Extracting tar file ~/.pot/xhyve/potMachine.tar.gz to ~/.pot/xhyve/")
	//untar potMachine.tar.gz into ~/.pot/xhyve
	r, err := os.Open(tarPath)
	if err != nil {
		fmt.Println("Error openning tar file with err: ", err)
	}
	extractTarGz(r, xhyveDirPath+"/")

	fmt.Println("==> Cleaning up ~/.pot/xhyve/")
	// delete file
	os.Remove(xhyveDirPath + "/metadata.json")

	fmt.Println("==> Enabeling nfs mountpoint")

	//Enable NFS on mac sudo nfsd enable
	enableNFS()

	chmodPrivateKey()

	//GET uid of current user
	UUID := os.Getuid()
	UUIDString := strconv.Itoa(UUID)

	//Edit NFS /etc/exports
	editNFSExports(UUIDString, potDirPath)

	//Check if runfile exists
	var runFile string
	if _, err := os.Stat(xhyveDirPath + "/runFreeBSD.sh"); os.IsNotExist(err) {
		//Create run file
		runFile = `#/bin/sh
UUID="-U efc58796-25ec-4003-b216-f20be8100685"
USERBOOT="` + potDirPath + `/xhyve/userboot.so"
IMG="` + potDirPath + `/xhyve/block0.img"
KERNELENV=""

MEM="-m 4G"
SMP="-c 2"
PCI_DEV="-s 0:0,hostbridge -s 31,lpc"
NET="-s 2:0,virtio-net"
IMG_HDD="-s 4:0,virtio-blk,$IMG"
LPC_DEV="-l com1,stdio"
ACPI="-A"

nohup xhyve $ACPI $MEM $SMP $PCI_DEV $LPC_DEV $NET $IMG_HDD $UUID -f fbsd,$USERBOOT,$IMG,"$KERNELENV" </dev/null >/dev/null 2>&1 &
`
		// Write runfile to ~/.pot/xhyve/runFreeBSD.sh
		xhyveRunFilePath := potDirPath + "/xhyve/runFreeBSD.sh"

		err = ioutil.WriteFile(xhyveRunFilePath, []byte(runFile), 0775)
		if err != nil {
			fmt.Println("ERROR: Error writting file to disk with err: \n", err)
			return
		}
	}

	//Initializa xhyve vm
	err = runXhyve()
	if err != nil {
		fmt.Println("Error creating xhyve vm with err: ", err)
		return
	}

	netcat()

	generateSSHConfig(potDirPath, xhyveIP)

	localIP := getLocalIP()
	fmt.Println("==> Local IP: ", localIP)

	//Restart sudo nfsd restart
	restartNFSService()

	mountNFSonVM(localIP)
}

// Get preferred outbound ip of this machine
func getLocalIP() (IP string) {
	vagrantDirPath := getVagrantDirPath()
	configSSHFile := vagrantDirPath + "sshConfig"

	termCmd := "ssh -F " + configSSHFile + " potMachine env | grep SSH_CLIENT | awk '{print $1}'"

	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	cmd.Wait()
	split := strings.Split(out.String(), "=")
	return split[1]
}

func mountNFSonVM(localIP string) {
	homeDir := getUserHome()
	command := "sudo mount " + localIP + ":" + homeDir + "/.pot /vagrant"
	redirectToVagrant([]string{command})
}

func generateSSHConfig(potDirPath string, xhyveIP string) {
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
	xhyvesshConfigFilePath := potDirPath + "/sshConfig"

	err := ioutil.WriteFile(xhyvesshConfigFilePath, []byte(sshConfig), 0775)
	if err != nil {
		log.Fatal("ERROR: Error writting file to disk with err: \n", err)
	}
}

func chmodPrivateKey() {
	privateKey, _ := os.UserHomeDir()
	privateKey = privateKey + "/.pot/xhyve/private_key"
	command := "chmod 600 " + privateKey
	cmd := exec.Command("bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error starting Xhyve VM with err: ", err)
	}
}

func restartNFSService() {

	command := "sudo nfsd restart"
	cmd := exec.Command("bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error starting Xhyve VM with err: ", err)
	}
	fmt.Println("==> nfsd restarted")
	cmd.Wait()
}

func runXhyve() error {
	potDirPath := getVagrantDirPath()
	xhyveDirPath := potDirPath + "/xhyve"
	termCmd := `sudo ` + xhyveDirPath + `/runFreeBSD.sh`
	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error starting Xhyve VM with err: ", err)
		return err
	}
	return nil
}

func editNFSExports(UUID string, potDir string) {
	termCmd := `sudo tee -a /etc/exports << 'EOF'
# POTMACHINE-Xhyve-Begin
` + potDir + ` -alldirs -mapall=` + UUID + `
# POTMACHINE-Xhyve-END
EOF`
	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error enabeling NFS with err: ", err)
		log.Fatal(err)
	}
	cmd.Wait()
}

func enableNFS() {

	termCmd := "sudo nfsd enable"
	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error enabeling NFS with err: ", err)
		log.Fatal(err)
	}
	cmd.Wait()
}

func netcat() {
	fmt.Println("==> Waiting for machine to start...")
	termCmd := "nc -l 1234"
	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error getting ip information from the VM with err: ", err)
		log.Fatal(err)
	}
	cmd.Wait()
	xhyveIP = out.String()
	fmt.Println("==> Machine started with ip: ", xhyveIP)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to download file")
	}
	defer resp.Body.Close()
	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		return errors.New("cannot determine progress without Content-Length")
	}
	size, err := strconv.ParseInt(contentLengthHeader, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "bad Content-Length %q", contentLengthHeader)
	}
	ctx := context.Background()
	r := progress.NewReader(resp.Body)
	go func() {
		progressChan := progress.NewTicker(ctx, r, size, 1*time.Second)
		for p := range progressChan {
			fmt.Printf("\r==> %v remaining...", p.Remaining().Round(time.Second))
		}
		fmt.Println("\r==> Download is completed")
	}()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return errors.Wrap(err, "failed to read body")
	}

	return nil

}

func extractTarGz(gzipStream io.Reader, xhyveDirPath string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if header.Name == "./" {
				break
			}
			if err := os.Mkdir(xhyveDirPath+header.Name, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(xhyveDirPath + header.Name)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			//fmt.Println("Ignoring file: ",header.Name)
		}

	}
}
