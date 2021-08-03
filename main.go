package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"encoding/json"
)

var warningmsg string
var pluginDataFile string = "plugins/plugins.json"

type PluginsInfo struct {
	Plugins []PluginInfo `json:"plugins"`
}

type PluginInfo struct {
	Name             string    `json:"name"`
	SecurityWarnings []Warning `json:"securityWarnings"`
}

type Warning struct {
	Versions []Version `json:"versions"`
	ID       string    `json:"id"`
	Message  string    `json:"message"`
	URL      string    `json:"url"`
	Active   bool      `json:"active"`
}

type Version struct {
	FirstVersion string `json:"firstVersion"`
	LastVersion  string `json:"lastVersion"`
}

func CheckSecurityWarnings() error {
	fmt.Println("check downloaded", "Data file is downloaded and extracted", hasDownloaded)

	jsonFile, err := os.Open(pluginDataFile)
	if err != nil {
		fmt.Println("check security warnings", "error while opening", err)
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("check security warnings", "error while reading", err)
		return err
	}
	var AllPluginData PluginsInfo
	err = json.Unmarshal(byteValue, &AllPluginData)
	if err != nil {
		fmt.Println("Check Security warnings", "error while converting to JSON", err)
		return err
	}
	fmt.Println("Checking through all the warnings")
	for _, plugin := range AllPluginData.Plugins {

		if pluginData, ispresent := pluginset[plugin.Name]; ispresent {
			for _, warning := range plugin.SecurityWarnings {
				for _, version := range warning.Versions {
					firstVersion := version.FirstVersion
					lastVersion := version.LastVersion
					if len(firstVersion) == 0 {
						firstVersion = "0" // setting default value in case of empty string
					}
					if len(lastVersion) == 0 {
						lastVersion = pluginData.Version // setting default value in case of empty string
					}

					fmt.Println("Security vulnerability in",plugin.Name,pluginData.Version)
	
				}
			}

		}

	}

	return nil
}

type PluginData struct {
	Version string
	Kind    string
}

var pluginset = make(map[string]PluginData)

func NewPluginData(Version string, Kind string) PluginData {
	var plugin PluginData
	plugin.Version = Version
	plugin.Kind = Kind
	return plugin
}

func Download(url string, filepath string) error {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := http.Client{
		Timeout: 100 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("Successfully Downloaded")
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil

}

func Extract(source string) error {
	reader, err := os.Open(source)

	if err != nil {
		return err
	}
	defer reader.Close()
	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	defer archive.Close()
	writer, err := os.Create(pluginDataFile)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err

}

var hasDownloaded bool = false

func RetrieveDataFile() {
	for {
		url := "https://ci.jenkins.io/job/Infra/job/plugin-site-api/job/generate-data/lastSuccessfulBuild/artifact/plugins.json.gzip"
		filepath := "plugins/plugins.json.gzip"
		fmt.Println("Retreiving file", "Host Url", url)
		err := Download(url, filepath)
		if err != nil {
			fmt.Println("Retrieving File", "Error while downloading", err)
			continue
		}

		fmt.Println("Retrieve File", "Downloaded", filepath)
		err = Extract(filepath)
		if err != nil {
			fmt.Println("Retreive File", "Error while extracting", err)
			continue
		}
		fmt.Println("Retreive File", "Successfully extracted", pluginDataFile)
		hasDownloaded = true
		time.Sleep(12 * time.Hour)

	}
}

func runningtime(s string) (string, time.Time) {
	log.Println("Start:	", s)
	return s, time.Now()
}

func track(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("End:	", s, "took", endTime.Sub(startTime))
}

func execute() {
	defer track(runningtime("execute"))
	CheckSecurityWarnings()
}

func first(val *bool) {
	time.Sleep(1 * time.Second)
	*val = true

}

func Validate(){
	pluginset["google-login"]=NewPluginData("1.2","user-defined")
	pluginset["mailer"]=NewPluginData("1.1","base")
	CheckSecurityWarnings()
}

func main() {
	go RetrieveDataFile()
	for {
		if hasDownloaded {
			Validate()
			break
		}

	}
}
