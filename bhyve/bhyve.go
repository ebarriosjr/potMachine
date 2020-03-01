package bhyve

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
	"time"

	"github.com/machinebox/progress"
	"github.com/pkg/errors"
)

var (
	bhyveIP string
)

//TODO:
// Prepare host machine:
// # kldload vmm
// # ifconfig tap0 create
// # sysctl net.link.tap.up_on_open=1
// ==> net.link.tap.up_on_open: 0 -> 1
// # ifconfig bridge0 create
// # ifconfig bridge0 addm vtnet0 addm tap0
// # ifconfig bridge0 up

// # sh /usr/share/examples/bhyve/vmrun.sh -c 4 -m 1024M -t tap0 -d ~/.pot/bhyve/block0.img potMachine
// /usr/share/examples/bhyve/vmrun.sh -c 2 -m 1024 -d /root/block0.img potMachine
// # bhyvectl --destroy --vm=potMachine

func initializeBhyve(verbose bool) {
	potDirPath := getVagrantDirPath()
	bhyveDirPath := potDirPath + "/bhyve"

	if _, err := os.Stat(bhyveDirPath); os.IsNotExist(err) {
		fmt.Println("==> Creating ~/.pot/bhyve directory")
		os.Mkdir(bhyveDirPath, 0775)
	}

	//Download from github respo bhyve.tar.zg -> ~/.pot/bhyve
	fileURL := "https://app.vagrantup.com/ebarriosjr/boxes/FreeBSD12.1-zfs/versions/0.0.1/providers/bhyve.box"
	tarPath := bhyveDirPath + "/potMachine.tar.gz"

	fmt.Println("==> Checking if tar file already exists on ~/.pot/bhyve/potMachine.tar.gz")
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		fmt.Println("==> Downloading tar file to ~/.pot/bhyve/potMachine.tar.gz")
		if err := downloadFile(tarPath, fileURL); err != nil {
			fmt.Println("Error downloading tar file from vagrant cloud with err: ", err)
			log.Fatal()
		}
	}

	fmt.Println("==> Extracting tar file ~/.pot/bhyve/potMachine.tar.gz to ~/.pot/bhyve/")
	//untar potMachine.tar.gz into ~/.pot/bhyve
	r, err := os.Open(tarPath)
	if err != nil {
		fmt.Println("Error openning tar file with err: ", err)
	}
	extractTarGz(r, bhyveDirPath+"/")

	fmt.Println("==> Cleaning up ~/.pot/bhyve/")
	// delete file
	os.Remove(bhyveDirPath + "/metadata.json")

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
	if _, err := os.Stat(bhyveDirPath + "/runFreeBSD.sh"); os.IsNotExist(err) {
		//Create run file
		runFile = `#/bin/sh
UUID="-U efc58796-25ec-4003-b216-f20be8100685"
USERBOOT="` + potDirPath + `/bhyve/userboot.so"
IMG="` + potDirPath + `/bhyve/block0.img"
KERNELENV=""

MEM="-m 4G"
SMP="-c 2"
PCI_DEV="-s 0:0,hostbridge -s 31,lpc"
NET="-s 2:0,virtio-net"
IMG_HDD="-s 4:0,virtio-blk,$IMG"
LPC_DEV="-l com1,stdio"
ACPI="-A"

nohup bhyve $ACPI $MEM $SMP $PCI_DEV $LPC_DEV $NET $IMG_HDD $UUID -f fbsd,$USERBOOT,$IMG,"$KERNELENV" </dev/null >/dev/null 2>&1 &
`
		// Write runfile to ~/.pot/bhyve/runFreeBSD.sh
		bhyveRunFilePath := potDirPath + "/bhyve/runFreeBSD.sh"

		err = ioutil.WriteFile(bhyveRunFilePath, []byte(runFile), 0775)
		if err != nil {
			fmt.Println("ERROR: Error writting file to disk with err: \n", err)
			return
		}
	}

	//Initializa bhyve vm
	err = runBhyve()
	if err != nil {
		fmt.Println("Error creating bhyve vm with err: ", err)
		return
	}

	netcat()

	generateSSHConfig(potDirPath, bhyveIP)

}

func generateSSHConfig(potDirPath string, bhyveIP string) {
	//generate sshConfig file
	sshConfig := `Host potMachine
		HostName ` + bhyveIP + `
		User vagrant
		Port 22
		UserKnownHostsFile /dev/null
		StrictHostKeyChecking no
		PasswordAuthentication no
		IdentityFile ~/.pot/bhyve/private_key
		IdentitiesOnly yes
		LogLevel FATAL
	  `
	bhyvesshConfigFilePath := potDirPath + "/sshConfig"

	err := ioutil.WriteFile(bhyvesshConfigFilePath, []byte(sshConfig), 0775)
	if err != nil {
		log.Fatal("ERROR: Error writting file to disk with err: \n", err)
	}
}

func chmodPrivateKey() {
	privateKey, _ := os.UserHomeDir()
	privateKey = privateKey + "/.pot/bhyve/private_key"
	command := "chmod 600 " + privateKey
	cmd := exec.Command("bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error starting bhyve VM with err: ", err)
	}
}

func runBhyve() error {
	potDirPath := getVagrantDirPath()
	bhyveDirPath := potDirPath + "/bhyve"
	termCmd := `sudo ` + bhyveDirPath + `/runFreeBSD.sh`
	cmd := exec.Command("bash", "-c", termCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error starting bhyve VM with err: ", err)
		return err
	}
	return nil
}

func editNFSExports(UUID string, potDir string) {
	termCmd := `sudo tee -a /etc/exports << 'EOF'
# POTMACHINE-bhyve-Begin
` + potDir + ` -alldirs -mapall=` + UUID + `
# POTMACHINE-bhyve-END
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
	bhyveIP = out.String()
	fmt.Println("==> Machine started with ip: ", bhyveIP)
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

func extractTarGz(gzipStream io.Reader, bhyveDirPath string) {
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
			if err := os.Mkdir(bhyveDirPath+header.Name, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(bhyveDirPath + header.Name)
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
