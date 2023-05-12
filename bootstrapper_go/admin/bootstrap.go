package admin

import (
	"bytes"
	"context"
	"io"
	"log"
	"mime/multipart"

	"bootstrapper_go/common"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/oauth2"
)

var (
	deviceUrl = "http://" + common.ApiGatewayHost + ":" + common.ApiGatewayPort + common.DeviceRoute + common.ApiVersionRoute
	kuiperUrl = "http://" + common.ApiGatewayHost + ":" + common.ApiGatewayPort + common.KuiperRoute
)

func Bootstrap(ctx context.Context) {
	log.Println("bootstraping the edge ...")
	uploadDeviceProfile(ctx)
	createDevice(ctx)
	createKuiperStream(ctx)
	createKuiperRule(ctx)
	log.Println("finished, the edge can receive data from sensor now")
}

func uploadDeviceProfile(ctx context.Context) {
	fileDir, err := os.Getwd()
	common.CheckFatalWithMessage(err, "canont get pwd")
	fileName := "sensorClusterDeviceProfile.yaml"
	filePath := path.Join(fileDir, "/", fileName)
	file, err := os.Open(filePath)
	common.CheckFatalWithMessage(err, "cannot open file")
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	common.CheckFatalWithMessage(err, "cannot create multipart")
	io.Copy(part, file)
	writer.Close()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: common.PreProcessToken(common.GatewayJwtTokenWithCRLF),
		TokenType:   "Bearer",
	}))
	resp, err := client.Post(deviceUrl+"/deviceprofile/uploadfile", writer.FormDataContentType(), body)
	common.CheckFatal(err)
	defer resp.Body.Close()
	log.Println("device profile payload:\n\t", "file: sensorClusterDeviceProfile.yaml")
	log.Println("uploaded device profile")
}

func createDevice(ctx context.Context) {
	log.Println("device payload:\n\t", common.DeviceJson)
	common.SendPostJson(ctx, common.GatewayJwtTokenWithCRLF, deviceUrl+"/device", common.DeviceJson, "created device configuration")
}

func createKuiperStream(ctx context.Context) {
	log.Println("Kuiper stream payload:\n\t", common.KuiperStreamJson)
	common.SendPostJson(ctx, common.GatewayJwtTokenWithCRLF, kuiperUrl+"/streams", common.KuiperStreamJson, "created kuiper stream")
}

func createKuiperRule(ctx context.Context) {
	log.Println("Kuiper rule payload:\n\t", common.KuiperRuleJson)
	common.SendPostJson(ctx, common.GatewayJwtTokenWithCRLF, kuiperUrl+"/rules", common.KuiperRuleJson, "created kuiper rule")
}
