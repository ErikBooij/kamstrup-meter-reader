package mqtt

import (
	"fmt"
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
	PublishInterval time.Duration      `json:"interval"`
}

type MQTTNotification struct {
	Register int16  `json:"register"`
	Topic    string `json:"topic"`
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

	go func() {
		ticker := time.NewTicker(config.PublishInterval)

		for {
			<-ticker.C

			for _, notification := range config.Notifications {
				registerValue := client.ReadRegisterWithRetry(notification.Register, 5, time.Millisecond*1000)

				if registerValue.Error() != nil {
					continue
				}

				mqttClient.Publish(notification.Topic, 0, true, fmt.Sprintf("%0.4f", registerValue.Value()))
			}
		}
	}()

	return nil
}
