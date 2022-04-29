package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"heat-meter-read-client/kamstrup"
	"heat-meter-read-client/mqtt"
	"heat-meter-read-client/server"
)

type config struct {
	Port string           `json:"serialPort"`
	MQTT *mqtt.MQTTConfig `json:"mqtt"`
}

func main() {
	config, err := loadConfig("meter-reader-config.json")

	if err != nil {
		fmt.Printf("Unable to load config. Stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}

	if config.Port == "" {
		fmt.Println("No serial port provided (using --port option), stopping.")

		os.Exit(1)

		return
	}

	kamstrupClient, err := kamstrup.CreateKamstrupClient(config.Port, time.Millisecond*500)

	if err != nil {
		fmt.Printf("Unable to create Kampstrup client (full error below), stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}

	defer kamstrupClient.ClosePort()

	if err := mqtt.PublishValuesOverMQTT(config.MQTT, kamstrupClient); err != nil {
		fmt.Printf("Unable to start MQTT publisher (full error below), stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}

	if err := server.CreateAndRunWebServer(kamstrupClient); err != nil {
		fmt.Printf("Unable to run web server (full error below), stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}
}

func loadConfig(filename string) (config, error) {
	var config config

	executable, err := os.Executable()

	if err != nil {
		return config, err
	}

	jsonFile, err := os.Open(path.Join(path.Dir(executable), filename))

	if err != nil {
		return config, err
	}

	defer jsonFile.Close()

	data, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)

	if err != nil {
		return config, err
	}

	return config, nil
}
