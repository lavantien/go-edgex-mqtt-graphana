package web

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bootstrapper_go/common"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

type Page struct {
	Title    string
	Interval int64
	Data     []DataRender
	Provider string
	Posted   []ThirdPartyRender
}

type ThirdPartyRender struct {
	Temperature int64 `json:"temperature,omitempty"`
	Humidity    int64 `json:"humidity,omitempty"`
}

type ThirdPartyData struct {
	Temperature float64 `json:"temperature,omitempty"`
	Humidity    float64 `json:"humidity,omitempty"`
}

type DataRender struct {
	Time        string `json:"time,omitempty"`
	Device      string `json:"device,omitempty"`
	Temperature int64  `json:"temperature,omitempty"`
	Humidity    int64  `json:"humidity,omitempty"`
}

type Payload struct {
	APIVersion string  `json:"apiVersion"`
	StatusCode int64   `json:"statusCode"`
	Events     []Event `json:"events"`
}

type Event struct {
	APIVersion  string    `json:"apiVersion"`
	ID          string    `json:"id"`
	DeviceName  string    `json:"deviceName"`
	ProfileName string    `json:"profileName"`
	SourceName  string    `json:"sourceName"`
	Origin      float64   `json:"origin"`
	Readings    []Reading `json:"readings"`
	Tags        Tags      `json:"tags"`
}

type Reading struct {
	ID           string  `json:"id"`
	Origin       float64 `json:"origin"`
	DeviceName   string  `json:"deviceName"`
	ResourceName string  `json:"resourceName"`
	ProfileName  string  `json:"profileName"`
	ValueType    string  `json:"valueType"`
	Value        string  `json:"value"`
}

type Tags struct {
	Gateway   string `json:"Gateway"`
	Latitude  string `json:"Latitude"`
	Longitude string `json:"Longitude"`
}

var (
	limit          = "-1"
	resourceUrl    = "http://" + common.ApiGatewayHost + ":" + common.ApiGatewayPort + common.ResourceRoute + common.ApiVersionRoute + "/event/all?limit=" + limit
	thirdPartyUrl  = common.ThirdPartyUrl
	page           Page
	payload        Payload
	thirdPartyData = make([]ThirdPartyData, 0)
)

func Start(ctx context.Context) {
	page.Title = "Demo UI"
	page.Interval = int64(common.Interval / time.Second)
	page.Data = []DataRender{{
		Time:        time.Now().Format(common.TimeFortmat),
		Device:      "Web Service",
		Temperature: 0,
		Humidity:    0,
	}}
	page.Provider = common.ThirdPartyUrl
	page.Posted = []ThirdPartyRender{{
		Temperature: 0,
		Humidity:    0,
	}}
	go func(ctx context.Context) {
		for {
			fetchData(ctx)
			time.Sleep(common.Interval)
		}
	}(ctx)
	go func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("recovered from the third party communication failure:\n\t", r)
			}
		}()
		time.Sleep(time.Second * 1)
		for {
			var blob ThirdPartyData
			if len(page.Data) != 0 {
				blob = ThirdPartyData{
					Temperature: float64(page.Data[0].Temperature),
					Humidity:    float64(page.Data[0].Humidity),
				}
			} else {
				blob = ThirdPartyData{
					Temperature: 0,
					Humidity:    0,
				}
			}
			postDataProvider(ctx, &blob)
			fetchDataProvider(ctx)
			time.Sleep(common.Interval)
		}
	}(ctx)
	log.Println("listening on localhost:4321 ...")
	http.HandleFunc("/", indexHandler)
	err := http.ListenAndServe(":4321", nil)
	common.CheckFatal(err)
}

func postDataProvider(ctx context.Context, blob *ThirdPartyData) {
	body := fmt.Sprintf(common.ThirdPartySchema, blob.Temperature, blob.Humidity)
	log.Println("third party payload:\n\t", body)
	common.SendPostJson(ctx, common.ThirdPartyJwtToken, thirdPartyUrl, body, "posted successfully to third-party "+page.Provider)
}

func fetchDataProvider(ctx context.Context) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: common.PreProcessToken(common.ThirdPartyJwtToken),
		TokenType:   "Bearer",
	}))
	resp, err := client.Get(thirdPartyUrl)
	common.CheckFatal(err)
	defer resp.Body.Close()
	thirdPartyData = make([]ThirdPartyData, 0)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&thirdPartyData)
	common.CheckFatal(err)
	fillThirdPartyData()
	log.Println("fetch data successfully from third-party " + page.Provider)
}

func fetchData(ctx context.Context) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: strings.TrimSuffix(common.GatewayJwtTokenWithCRLF, "\r\n"),
		TokenType:   "Bearer",
	}))
	resp, err := client.Get(resourceUrl)
	common.CheckFatal(err)
	defer resp.Body.Close()
	payload = Payload{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&payload)
	common.CheckFatal(err)
	fillData()
	log.Println("fetch data successfully from edgex")
}

func fillData() {
	data := []DataRender{}
	for _, event := range payload.Events {
		var item DataRender
		item.Device = event.DeviceName
		item.Time = time.Unix(0, int64(event.Origin)).Format(common.TimeFortmat)
		for _, reading := range event.Readings {
			switch reading.ResourceName {
			case "temperature":
				temp, err := strconv.ParseInt(reading.Value, 10, 64)
				common.CheckFatal(err)
				item.Temperature = temp
			case "humidity":
				temp, err := strconv.ParseInt(reading.Value, 10, 64)
				common.CheckFatal(err)
				item.Humidity = temp
			}
		}
		data = append(data, item)
	}
	page.Data = data
}

func fillThirdPartyData() {
	data := []ThirdPartyRender{}
	for _, event := range thirdPartyData {
		var item ThirdPartyRender
		item.Temperature = int64(event.Temperature)
		item.Humidity = int64(event.Humidity)
		data = append(data, item)
	}
	common.ReverseAny(data)
	page.Posted = data
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tPath := "/web/ui.gohtml"
	t := template.Must(template.ParseFiles(tPath))
	err := t.Execute(w, page)
	common.CheckFatal(err)
	log.Println("UI serve successfully")
}
