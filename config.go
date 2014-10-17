package main

import (
	"io/ioutil"
	"os"
	"text/template"
	"fmt"
	"encoding/json"
	"sync"
)

var LocalFilename string
var TemplateFile string
var ConfigFile string


func SetFileName(c string) error {
	LocalFilename = c
	return nil
}

func SetTemplateFileName(c string) error {
	TemplateFile = c
	return nil
}

func SetConfigFileName(c string) error {
	ConfigFile = c
	return nil
}

// updates the weight of a server of a specific backend with a new weight
func UpdateWeightInConfig(backend string, server string, weight int, config *Config) error {

	config.Mutex.RLock()
	defer config.Mutex.RUnlock()

	for _, be := range config.Backends {
		fmt.Printf(be.Name)
		if be.Name == backend {
			for _, srv := range be.BackendServers {
				if srv.Name == server {
					srv.Weight = weight
				}
			}
		}
	}

	err := WriteConfigToDisk(config)
	return err
}

// Render a config object to a HAproxy config file
func RenderConfig(config *Config) error {
	f, err := ioutil.ReadFile(TemplateFile)
	if err != nil {
		return err
	}

	fp, err := os.OpenFile(ConfigFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()

	config.Mutex.RLock()
	defer config.Mutex.RUnlock()

	// before rendering, commit config to disk

	err = WriteConfigToDisk(config)
	if err != nil {

		return err
	}

	t := template.Must(template.New(TemplateFile).Parse(string(f)))
	err = t.Execute(fp, &config)
	if err != nil {
		return err
	}

	return nil
}

func GetConfigFromDisk() (*Config, error) {
	c := ConfigObj
	s, err := ioutil.ReadFile(LocalFilename)
	if err != nil {
		return c,err
	}
	err = json.Unmarshal(s, &ConfigObj)
	if err != nil {
		fmt.Println("Error parsing JSON")
		return c,err
	}

	ConfigObj.Mutex = new(sync.RWMutex)
	return ConfigObj, err
}

func WriteConfigToDisk(config *Config) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(LocalFilename, b, 0666)
	if err != nil {
		return err
	}
	return nil
}


func RenderLocalProxyConfig(payload []byte, conf *Config) {

	json.Unmarshal(payload, conf)
	err := RenderConfig(ConfigObj)
	if err != nil {
		log.Error("Error rendering config file")
		return
	} else {
		err = Reload()
		if err != nil {
			log.Error("Error reloading the HAproxy configuration")
			return
		} else {
			log.Info("Successfully reloaded HAproxy configuration")
		}

	}

}
