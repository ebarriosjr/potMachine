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
	//fmt.Println("fileName: ", fileName)

	tagXZ := *tag + ".xz"
	i := strings.LastIndex(tagXZ, ":")

	tagXZ = tagXZ[:i] + strings.Replace(tagXZ[i:], ":", "_", 1)
	//fmt.Println("tagXZ: ", tagXZ)

	//fmt.Println("URL: ", tagXZ, " \nFilename: ", fileName)
	fmt.Println("==> Pushing " + name + " version: " + version + " xz file...")
	response := sendPutRequest(tagXZ, fileName, "xz")
	if response.StatusCode != 201 {
		fmt.Println("ERROR: Error executing command with http response: ", response.StatusCode)
		fmt.Println("ERROR: Status: ", response)
	}

	shaName := fileName + ".skein"
	//fmt.Println("shaName: ", shaName)

	tagSKEIN := tagXZ + ".skein"

	//fmt.Println("URL: ", tagSKEIN, " \nFilename: ", shaName)
	fmt.Println("==> Pushing " + name + " version: " + version + " xz.skein file...")
	response = sendPutRequest(tagSKEIN, shaName, "skein")
	if response.StatusCode != 201 {
		fmt.Println("ERROR: Error executing command with http response: ", response.StatusCode)
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
		fmt.Println("Error opening file with err: ", err)
		return nil
	}
	defer file.Close()

	request, err := http.NewRequest("PUT", url, file)
	if err != nil {
		fmt.Println("Error creating http request with err:", err)
		return nil
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	auth, err := getBasicAuth(url)
	if err != nil {
		fmt.Println("Trying with no user data")
	} else {
		request.Header.Add("Authorization", "Basic "+auth)
	}

	response, err := client.Do(request)

	if err != nil {
		fmt.Println("Error client.Do with err: ", err)
		return nil
	}
	defer response.Body.Close()

	_, err = ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println("Error reading response with err: ", err)
		return nil
	}

	return response
}

func getBasicAuth(url string) (basic string, err error) {
	// Adding Basic auth logic
	homeDir := getUserHome()
	if homeDir == "" {
		fmt.Println("ERROR: Error getting user home directory info...")
		return "", nil
	}
	configFilePath := homeDir + "/.pot/config.json"

	var configFile ConfigFile
	_, err = os.Stat(configFilePath)
	if err == nil {
		//File exists
		dat, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			fmt.Println("ERROR: Unable to read config.json file with err: \n", err)
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
