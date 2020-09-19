package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/ttacon/gofourit"
	"github.com/ztrue/shutdown"
)

var (
	id = flag.String("id", "", "ID for this process")
)

func main() {
	flag.Parse()
	if len(*id) == 0 {
		log.Println("must provide -id flag")
		os.Exit(1)
	}

	rCron := gofourit.New(
		gofourit.NewRedisRemoteSource(
			redis.NewClient(
				&redis.Options{
					Network: "tcp",
					Addr:    "127.0.0.1:6379",
				},
			),
		),
	)

	shHandler := shutdown.New()
	shHandler.Add(func() {
		rCron.Stop()
	})

	rCron.AddFunc("* * * * *", "print-it", func() {
		fmt.Printf("running `print-it` from: %q\n", *id)
	})

	rCron.AddFunc("* * * * *", "say-hello", func() {
		fmt.Printf("hello from: %q\n", *id)
	})
	rCron.Start()

	log.Println("up and running...")
	shHandler.Listen(syscall.SIGTERM)
}
