package mqtt

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"heat-meter-read-client/kamstrup"
)

type MQTTConfig struct {
	Host            string             `json:"host"`
	Port            uint16             `json:"port"`
	User            string             `json:"user"`
	Pass            string             `json:"pass"`
	Notifications   []MQTTNotification `json:"notifications"`
	PublishInterval int                `json:"interval"`
	Prefix          string             `json:"prefix"`
}

type MQTTNotification struct {
	Register int16  `json:"register"`
	ID       string `json:"id"`
}

func PublishValuesOverMQTT(config *MQTTConfig, client kamstrup.KamstrupClient) error {
	mqttBroker := fmt.Sprintf("tcp://%s:%d", config.Host, config.Port)

	opts := mqtt.NewClientOptions().
		AddBroker(mqttBroker).
		SetUsername(config.User).
		SetPassword(config.Pass).
		SetAutoReconnect(true).
		SetClientID("kamstrup-meter-reader")

	mqttClient := mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	var topicFn = func(readingIdentifier string) func(string) string {
		return func(path string) string {
			return fmt.Sprintf("%s/%s/%s", strings.Trim(config.Prefix, "/"), strings.Trim(readingIdentifier, "/"), strings.Trim(path, "/"))
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(config.PublishInterval))

		for {
			<-ticker.C

			for _, notification := range config.Notifications {
				topic := topicFn(notification.ID)

				registerValue := client.ReadRegisterWithRetry(notification.Register, 5, time.Millisecond*1000)

				mqttClient.Publish(topic("/meta/latest-attempt-at"), 0, true, time.Now().Format(time.RFC3339))

				if registerValue.Error() != nil {
					mqttClient.Publish(topic("/meta/latest-error"), 0, true, registerValue.Error().Error())
					mqttClient.Publish(topic("/meta/latest-error-at"), 0, true, time.Now().Format(time.RFC3339))

					continue
				}

				mqttClient.Publish(topic("/value"), 0, true, fmt.Sprintf("%0.4f", registerValue.Value()))
				mqttClient.Publish(topic("/meta/latest-reading-at"), 0, true, time.Now().Format(time.RFC3339))
			}
		}
	}()

	return nil
}
