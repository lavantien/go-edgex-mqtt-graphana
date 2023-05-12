package common

import (
	_ "embed"
)

var (
	//go:embed cfg_thirdparty_host
	ThirdPartyHost string
	//go:embed cfg_thirdparty_port
	ThirdPartyPort string
	//go:embed cfg_thirdparty_jwt_token
	ThirdPartyJwtToken string
	ThirdPartyEndpoint = ThirdPartyHost + ":" + ThirdPartyPort
	ContentType        = "application/json"
)
