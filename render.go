package main

import (
	"io/ioutil"
	"os"
	"text/template"
)

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
