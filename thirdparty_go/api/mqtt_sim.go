package api

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Panicln(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername("mqtt")
	password := ""
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func RunMqttSender() {
	uri, err := url.Parse("tcp://localhost:1883")
	if err != nil {
		log.Panicln(err)
	}
	client := connect("pub", uri)
	timer := time.NewTicker(1 * time.Second)
	for t := range timer.C {
		fmt.Println(t)
		var min int64 = 0
		var max int64 = 100
		var random int64 = (rand.Int63n(max-min) + min)
		nsec := time.Now().UnixNano()
		payload := `{
			"id": ` + uuid.NewString() + `,
			"created": ` + strconv.FormatUint(uint64(nsec), 10) + `,
			"origin": ` + strconv.FormatUint(uint64(nsec), 10) + `,
			"device": "Temp_and_Humidity_sensor_cluster_01",
			"name": "temperature",
			"value": ` + strconv.FormatUint(uint64(random), 10) + `,
			"valueType": "Int64"
		}`
		client.Publish("mqtt_consumer", 0, false, payload)
		random = (rand.Int63n(max-min) + min)
		nsec = time.Now().UnixNano()
		payload = `{
			"id": ` + uuid.NewString() + `,
			"created": ` + strconv.FormatUint(uint64(nsec), 10) + `,
			"origin": ` + strconv.FormatUint(uint64(nsec), 10) + `,
			"device": "Temp_and_Humidity_sensor_cluster_01",
			"name": "humidity",
			"value": ` + strconv.FormatUint(uint64(random), 10) + `,
			"valueType": "Int64"
		}`
		client.Publish("mqtt_consumer", 0, false, payload)
	}
}

func Prod() bool {
	return false
}
