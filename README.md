gofourit
========

# Motivation
You need to run a single function on a cron-style schedule, but you need to
ensure that it only runs once across a fleet of tasks or servers. As such, this
library is more of an opinionated take on how to do this as opposed to a general
solution.

# Installation

The usual manner of grabbing the go package:
```
go get github.com/ttacon/gofourit
```

# Usage

The `examples/` folder has more examples, but basic usage looks as follows:

```go
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
```

# How it works
## Unique key per function
`gofourit` works by wrapping each registered function in a function that
attempts to acquire a lock based on the given name. If the function gets the
lock, it'll run the function, if it doesn't it returns immediately.

Because we use a unique key per function, this means functions will be run by
whichever owning process grabs the lock.

## Compared to `cron-cluster` in nodejs

This library takes inspiration in spirit from `cron-cluster` in Node.js, but
differs in one primary way. `cron-cluster` uses an underlying algorithm to
identify who should be the `leader` for the given instantiation of that
`CronJob` (it uses `redis-leader` for this `leader` identification). This means
that if you register multiple cron jobs, they will all run on whichever task
is currently the `leader`.

`gofourit` instead takes the stance to share this load out, instead of
potentially isolating all cron jobs to run on the same task.

# Future enhancements

 - Better logging options
 - Better lock configuration options (TTLs, retries, maintaining a lock)
 - Further remote source implementations (e.g. DynamoDB)
