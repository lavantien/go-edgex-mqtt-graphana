package main

import (
	"context"
	"log"
	"thirdparty_go/api"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	time.Sleep(time.Second * 1)
	ctx := context.Background()
	api.ListenHttpRest(ctx)
}
