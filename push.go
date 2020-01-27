package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cheggaaa/pb/v3"
)

func pushPot(tag *string) {
	// curl -u myUser:myPA$$wd! -X PUT "http://localhost:8081/artifactory/my-repository/my/new/artifact/directory/file.txt" -T Desktop/myNewFile.txt
	vagrantDirPath := getVagrantDirPath()

	if !strings.Contains(*tag, "http") {
		tag2 := "https://" + *tag
		*tag = tag2
	}

	// Get the info after the last / should be potName:Version
	splitName := strings.Split(*tag, "/")
	length := len(splitName) - 1
	namePlusVersion := splitName[length]
	splitNamePlusVersion := strings.Split(namePlusVersion, ":")

	name := splitNamePlusVersion[0]
	version := splitNamePlusVersion[1]

	fileName := vagrantDirPath + "exports/" + name + "_" + version + ".xz"
	//log.Info("fileName: ", fileName)

	tagXZ := *tag + ".xz"
	i := strings.LastIndex(tagXZ, ":")

	tagXZ = tagXZ[:i] + strings.Replace(tagXZ[i:], ":", "_", 1)
	//log.Info("tagXZ: ", tagXZ)

	//log.Info("URL: ", tagXZ, " \nFilename: ", fileName)
	fmt.Println("==> Pushing " + name + " version: " + version + " xz file...")
	response := sendPutRequest(tagXZ, fileName, "xz")
	if response.StatusCode != 201 {
		log.Error("Error executing command with http response: ", response.StatusCode)
		log.Error("Status: ", response)
	}

	shaName := fileName + ".skein"
	//log.Info("shaName: ", shaName)

	tagSKEIN := tagXZ + ".skein"

	//log.Info("URL: ", tagSKEIN, " \nFilename: ", shaName)
	fmt.Println("==> Pushing " + name + " version: " + version + " xz.skein file...")
	response = sendPutRequest(tagSKEIN, shaName, "skein")
	if response.StatusCode != 201 {
		log.Error("Error executing command with http response: ", response.StatusCode)
	}
}

func progressBar(buffer *bytes.Buffer) {
	count := buffer.Len()
	total := count
	// create and start new bar
	bar := pb.SuperSimple.Start(count)

	var current int
	for i := 0; i < count; i++ {
		current = total - buffer.Len()
		bar.SetCurrent(int64(current))
		time.Sleep(time.Second)
	}
	bar.Finish()

}

func sendPutRequest(url string, filename string, filetype string) *http.Response {
	file, err := os.Open(filename)

	if err != nil {
		log.Fatal("Error opening file with err: ", err)
	}
	defer file.Close()

	request, err := http.NewRequest("PUT", url, file)
	if err != nil {
		log.Fatal("Error creating http request with err:", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	auth, err := getBasicAuth(url)
	if err != nil {
		log.Info("Trying with no user data")
	} else {
		request.Header.Add("Authorization", "Basic "+auth)
	}

	response, err := client.Do(request)

	if err != nil {
		log.Fatal("Error client.Do with err: ", err)
	}
	defer response.Body.Close()

	_, err = ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal("Error reading response with err: ", err)
	}

	return response
}

func getBasicAuth(url string) (basic string, err error) {
	// Adding Basic auth logic
	homeDir := getUserHome()
	if homeDir == "" {
		log.Error("Error getting user home directory info...")
		return "", nil
	}
	configFilePath := homeDir + "/.pot/config.json"

	var configFile ConfigFile
	_, err = os.Stat(configFilePath)
	if err == nil {
		//File exists
		dat, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			log.Error("Unable to read config.json file with err: \n", err)
			return "", err
		}
		err = json.Unmarshal(dat, &configFile)

	} else if os.IsNotExist(err) {
		fmt.Println("No authetication set for this endpoint. \nTrying to push without authentication.\nIf this doesn't work please run pot login [SERVER]")
		return "", err
	}

	server := strings.Split(url, "/")
	if auth, ok := configFile.AuthConfigs[server[2]]; ok {
		return auth.Auth, nil
	}

	return "", nil
}
