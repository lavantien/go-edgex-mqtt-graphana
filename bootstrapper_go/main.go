package main

import (
	"bootstrapper_go/admin"
	"bootstrapper_go/sensor"
	"bootstrapper_go/web"
	"context"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	time.Sleep(time.Second * 1)
	ctx := context.Background()
	admin.Bootstrap(ctx)
	go sensor.Run(ctx)
	web.Start(ctx)
}
