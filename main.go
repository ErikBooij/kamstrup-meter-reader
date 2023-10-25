package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"heat-meter-read-client/kamstrup"
	"heat-meter-read-client/mqtt"
	"heat-meter-read-client/server"
)

type config struct {
	Devices map[string]device `json:"devices"`
	MQTT    *mqtt.MQTTConfig  `json:"mqtt"`
}

type device struct {
	Port string `json:"serialPort"`
}

func main() {
	log.Println("Starting meter reader...")

	config, err := loadConfig("meter-reader-config.json")

	if err != nil {
		log.Printf("Unable to load config. Stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}

	log.Println("Config loaded.")

	clients := make(map[string]kamstrup.KamstrupClient)

	for name, device := range config.Devices {
		clients[name] = kamstrup.CreateKamstrupClient(device.Port, time.Millisecond*500)

		if err != nil {
			log.Printf("Unable to create Kampstrup client for device %s on port %s (full error below), stopping.\n\n%s\n", name, device.Port, err)

			os.Exit(1)

			return
		}
	}

	if err := mqtt.PublishValuesOverMQTT(config.MQTT, clients); err != nil {
		log.Printf("Unable to start MQTT publisher (full error below), stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}

	log.Println("MQTT publisher started.")
	log.Println("Starting web server...")

	if err := server.CreateAndRunWebServer(clients, config.MQTT.Notifications); err != nil {
		log.Printf("Unable to run web server (full error below), stopping.\n\n%s\n", err)

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
