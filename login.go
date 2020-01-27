package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

func loginPot(server []string, user *string, pass *string, inline *bool) {
	if *user == "" && *pass == "" {
		*inline = true
		fmt.Print("Enter Username: ")
		reader := bufio.NewReader(os.Stdin)
		username, _ := reader.ReadString('\n')
		user = &username
	}

	if *inline {
		//Get the pass from stdin
		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			fmt.Println("Error: Error inputing password with err: ", err)
			return
		}
		password := string(bytePassword)

		pass = &password
	}

	if len(server) == 0 || *user == "" || (*pass == "" && *inline == false) {
		fmt.Println("Usage: pot login [OPTIONS] [SERVER] \n\nLog in a webserver registry\n\nOptions:\n  -p string -- Password\n  -u string -- Username")
		return
	}

	// Open ~/.pot/config.json if it exists, if not create it
	// Add username and password to config.json if
	client := http.Client{Timeout: 5 * time.Second}

	var prefixServer string
	if !strings.Contains(server[0], "http") {
		prefix := "https://"
		prefixServer = prefix + server[0]
	} else {
		prefixServer = server[0]
	}

	req, err := http.NewRequest("GET", prefixServer, nil)
	if err != nil {
		fmt.Println("Error: Unable to create request with err: ", err)
		return
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(*user, *pass))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: Unable to comunicate with endpoint with err: \n", err)
		return
	}
	homeDir := getUserHome()
	if homeDir == "" {
		fmt.Println("Error: Error getting user home directory info..")
		return
	}
	configFilePath := homeDir + "/.pot/config.json"

	if resp.StatusCode == 200 {
		//Read FILE and write json
		var configFile ConfigFile
		_, err := os.Stat(configFilePath)
		if err == nil {
			//File exists
			dat, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				fmt.Println("Error: Unable to read config.json file with err: \n", err)
				return
			}
			err = json.Unmarshal(dat, &configFile)

		} else if os.IsNotExist(err) {
			AuthConfigs := make(map[string]AuthConfig)
			configFile.AuthConfigs = AuthConfigs
		}

		var auth AuthConfig
		auth.Auth = basicAuth(*user, *pass)

		configFile.AuthConfigs[server[0]] = auth

		//newFlavorByte, _ := json.Marshal(configFile)
		newFlavorByte, _ := json.MarshalIndent(configFile, "", "    ")
		err = ioutil.WriteFile(configFilePath, newFlavorByte, 0644)
		if err != nil {
			fmt.Println("Error: Error writting file to disk with err: \n", err)
			return
		}

	} else if resp.StatusCode == 401 {
		fmt.Println("Error: Not Authorized")
		return
	} else {
		fmt.Println("Error: Error reaching server with err: \n", err)
		return
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
