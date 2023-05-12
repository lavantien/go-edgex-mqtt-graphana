package common

import (
	"bytes"
	"context"
	_ "embed"
	"log"
	"reflect"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

var (
	//go:embed cfg_time_format
	TimeFortmat string
	//go:embed cfg_gateway_host
	ApiGatewayHost string
	//go:embed cfg_gateway_port
	ApiGatewayPort string
	//go:embed cfg_thirdparty_host
	ThirdPartyHost string
	//go:embed cfg_thirdparty_port
	ThirdPartyPort string
	//go:embed cfg_thirdparty_jwt_token
	ThirdPartyJwtToken string
	//go:embed gateway_jwt_token
	GatewayJwtTokenWithCRLF string
	//go:embed root_token.json
	RootTokenJsonWithCRLF string
	Interval              = time.Second * 5
	Provider              = "/data"
	ThirdPartyUrl         = "http://" + ThirdPartyHost + ":" + ThirdPartyPort + Provider
	DevicePort            = "59881"
	ResourcePort          = "59880"
	KuiperPort            = "59720"
	ApiVersionRoute       = "/api/v2"
	DeviceRoute           = "/core-metadata"
	ResourceRoute         = "/core-data"
	KuiperRoute           = "/rules-engine"
	DeviceProfileFilePath = "sensorClusterDeviceProfile.yaml"
	PayloadType           = "application/json"
	DeviceJson            = `
	[
		{
			"apiVersion": "v2",
			"device": {
				"name": "Temp_and_Humidity_sensor_cluster_01",
				"description": "Raspberry Pi sensor cluster",
				"adminState": "UNLOCKED",
				"operatingState": "UP",
				"labels": [
					"Humidity sensor",
					"Temperature sensor",
					"DHT11"
				],
				"location": "{lat:45.45,long:47.80}",
				"serviceName": "device-rest",
				"profileName": "SensorCluster",
				"protocols": {
					"example": {
						"host": "dummy",
						"port": "1234",
						"unitID": "1"
					}
				}
			}
		}
	]
	`
	KuiperStreamJson = `
	{
		"sql": "create stream MQTT_TOPIC() WITH (FORMAT=\"JSON\", TYPE=\"edgex\")"
	}
	`
	KuiperRuleJson = `
	{
		"id": "mqtt_export_rule",
		"sql": "SELECT * FROM MQTT_TOPIC",
		"actions": [
			{
				"mqtt": {
					"server": "tcp://mosquitto:1883",
					"topic": "MQTT_TOPIC",
					"username": "someuser",
					"password": "somepassword",
					"clientId": "someclientid"
				}
			},
			{
				"log": {}
			}
		]
	}
	`
	ThirdPartySchema = `
	{
		"temperature": %f,
		"humidity": %f
	}
	`
	ResourceSchema = `
	{
		"apiVersion": "v2",
		"event": {
			"apiVersion": "v2",
			"deviceName": "Temp_and_Humidity_sensor_cluster_01",
			"profileName": "SensorCluster",
			"sourceName": "%s",
			"id": "%s",
			"origin": %d,
			"tags": {
				"Gateway": "MarsColony-12345",
				"Latitude": "45.450000",
				"Longitude": "47.470000"
			},
			"readings": [
				{
					"deviceName": "Temp_and_Humidity_sensor_cluster_01",
					"resourceName": "temperature",
					"profileName": "SensorCluster",
					"id": "%s",
					"origin": %d,
					"valueType": "Int64",
					"value": "%d"
				},
				{
					"deviceName": "Temp_and_Humidity_sensor_cluster_01",
					"resourceName": "humidity",
					"profileName": "SensorCluster",
					"id": "%s",
					"origin": %d,
					"valueType": "Int64",
					"value": "%d"
				}
			]
		}
	}
	`
	TemperatureUpperBound int64 = 100
	TemperatureLowerBound int64 = -50
	HumidityUpperBound    int64 = 100
	HumidityLowerBound    int64 = 0
)

func CheckFatal(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func CheckFatalWithMessage(err error, message string) {
	if err != nil {
		log.Panicln(message, "-", err)
	}
}

func SendPostJson(ctx context.Context, jwtToken string, url string, payload string, logMessage string) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: PreProcessToken(jwtToken),
		TokenType:   "Bearer",
	}))
	body := bytes.NewBufferString(payload)
	resp, err := client.Post(url, PayloadType, body)
	CheckFatal(err)
	defer resp.Body.Close()
	log.Println(logMessage)
}

func PreProcessToken(embeddedToken string) string {
	n := len(embeddedToken)
	var token string
	if n > 0 && embeddedToken[n-1] == '\n' {
		if embeddedToken[n-2] == '\r' {
			token = strings.TrimSuffix(embeddedToken, "\r\n")
		} else {
			token = strings.TrimSuffix(embeddedToken, "\n")
		}
	} else {
		token = embeddedToken
	}
	return token
}

func ReverseAny(s interface{}) {
    n := reflect.ValueOf(s).Len()
    swap := reflect.Swapper(s)
    for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
        swap(i, j)
    }
}