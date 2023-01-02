package server

import (
	"encoding/json"
	"fmt"
	"heat-meter-read-client/mqtt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"heat-meter-read-client/kamstrup"
)

type requestParameters struct {
	reg     int16
	retries int
	backoff time.Duration
}

type registeredSensorValue struct {
	Error  string  `json:"error,omitempty"`
	Name   string  `json:"name"`
	RegDec string  `json:"regDec"`
	RegHex string  `json:"regHex"`
	Value  float64 `json:"value"`
}

func CreateAndRunWebServer(kamstrupClient kamstrup.KamstrupClient, notifications []mqtt.MQTTNotification) error {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		values := make([]registeredSensorValue, len(notifications))

		i := 0

		for i < len(notifications) {
			reading := kamstrupClient.ReadRegister(notifications[i].Register)

			errorValue := ""

			if reading.Error() != nil {
				errorValue = fmt.Sprintf("%s", reading.Error())
			}

			values[i] = registeredSensorValue{
				Error:  errorValue,
				Name:   notifications[i].ID,
				RegDec: fmt.Sprintf("%d", notifications[i].Register),
				RegHex: fmt.Sprintf("%x", notifications[i].Register),
				Value:  reading.Value(),
			}

			i++
		}

		sort.SliceStable(values, func(i, j int) bool {
			return strings.Compare(values[i].Name, values[j].Name) < 0
		})

		responseBody, _ := json.Marshal(map[string]interface{}{
			"notifications": values,
		})

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(responseBody)
	})

	http.HandleFunc("/read", func(writer http.ResponseWriter, request *http.Request) {
		params := extractParameters(request)

		registerValue := kamstrupClient.ReadRegisterWithRetry(params.reg, params.retries, params.backoff)

		if registerValue.Error() != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Unable to query meter"))

			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(fmt.Sprintf("%0.2f", registerValue.Value())))
	})

	return http.ListenAndServe(":80", nil)
}

func extractParameters(request *http.Request) requestParameters {
	return requestParameters{
		reg:     extractRegisterNumber(request),
		retries: extractRetries(request),
		backoff: extractBackoff(request),
	}
}

func extractBackoff(request *http.Request) time.Duration {
	backoffString := request.URL.Query().Get("backoff")

	if backoffString == "" {
		return 0
	}

	backoffInt, err := strconv.ParseInt(backoffString, 10, 8)

	if err != nil {
		return 0
	}

	return time.Millisecond * time.Duration(backoffInt)
}

func extractRegisterNumber(request *http.Request) int16 {
	regString := request.URL.Query().Get("register")

	if regString == "" {
		return 0
	}

	registerInt, err := strconv.ParseInt(regString, 10, 16)

	if err != nil {
		return 0
	}

	return int16(registerInt)
}

func extractRetries(request *http.Request) int {
	retriesString := request.URL.Query().Get("retries")

	if retriesString == "" {
		return 0
	}

	retriesInt, err := strconv.ParseInt(retriesString, 10, 8)

	if err != nil {
		return 0
	}

	return int(retriesInt)
}
