package main

//ConfigFile struct
type ConfigFile struct {
	AuthConfigs map[string]AuthConfig `json:"auths"`
}

//AuthConfig struct
type AuthConfig struct {
	Auth string `json:"auth,omitempty"`
}

type copy struct {
	source      string
	destination string
}

type env struct {
	variable string
	value    string
}

//Config struct
type Config struct {
	Editor string
	VMType string
	IP     string
	Memory string
	Cpus   string
}
