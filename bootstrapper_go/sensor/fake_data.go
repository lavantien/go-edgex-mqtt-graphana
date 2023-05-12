package sensor

import (
	"bootstrapper_go/common"
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	profileName       = "SensorCluster"
	deviceName        = "Temp_and_Humidity_sensor_cluster_01"
	resourceUrl       = "http://" + common.ApiGatewayHost + ":" + common.ApiGatewayPort + common.ResourceRoute + common.ApiVersionRoute + "/event/" + profileName + "/" + deviceName
	rd                = rand.New(rand.NewSource(time.Now().UnixNano()))
	payloadType       = common.PayloadType
	temperature int64 = 28
	humidity    int64 = 60
)

func Run(ctx context.Context) {
	for {
		temperature, humidity = generateSensorData(temperature, humidity)
		sendData(ctx, "combine", temperature, humidity)
		log.Printf("sent: temperature %dC, humidity %d%%", temperature, humidity)
		time.Sleep(common.Interval)
	}
}

func sendData(ctx context.Context, resource string, data1 int64, data2 int64) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: common.PreProcessToken(common.GatewayJwtTokenWithCRLF),
		TokenType:   "Bearer",
	}))
	eventUuid := uuid.NewString()
	readingUuid1 := uuid.NewString()
	readingUuid2 := uuid.NewString()
	timestamp := time.Now().UnixNano()
	payload := fmt.Sprintf(common.ResourceSchema, resource, eventUuid, timestamp, readingUuid1, timestamp, data1, readingUuid2, timestamp, data2)
	body := bytes.NewBufferString(payload)
	resp, err := client.Post(resourceUrl+"/"+resource, payloadType, body)
	common.CheckFatal(err)
	defer resp.Body.Close()
	log.Println("resource payload:\n\t", payload)
}

func generateSensorData(inputTemperature, inputHumidity int64) (int64, int64) {
	temperatureMin := inputTemperature - 2
	if temperatureMin < common.TemperatureLowerBound {
		temperatureMin = common.TemperatureLowerBound
	}
	temperatureMax := inputTemperature + 2
	if temperatureMax > common.TemperatureUpperBound {
		temperatureMax = common.TemperatureUpperBound
	}
	humidityMin := inputHumidity - 4
	if humidityMin < common.HumidityLowerBound {
		humidityMin = common.HumidityLowerBound
	}
	humidityMax := inputHumidity + 4
	if humidityMax > common.HumidityUpperBound {
		humidityMax = common.HumidityUpperBound
	}
	temperature := rd.Int63n(temperatureMax-temperatureMin+1) + temperatureMin
	humidity := rd.Int63n(humidityMax-humidityMin+1) + humidityMin
	return temperature, humidity
}
