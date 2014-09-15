package main

import (
"io/ioutil"
"os"
"text/template"
"fmt"
)

// updates the weight of a server of a specific backend with a new weight
func UpdateWeightInConfig(backend string, server string, weight int, config *Config) *Config {

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
	return config
}

// Render a config object to a HAproxy config file
func RenderConfig(outFile string, templateFile string, config *Config) error {
	f, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return err
	}

	fp, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()

	config.Mutex.RLock()
	defer config.Mutex.RUnlock()
	t := template.Must(template.New(templateFile).Parse(string(f)))
	err = t.Execute(fp, &config)
	if err != nil {
		return err
	}

	return nil
}
