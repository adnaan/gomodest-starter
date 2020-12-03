package main

import (
	"context"
	"flag"
	"fmt"
	"gomodest/pkg"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/valve"
)

func main() {
	// Our graceful valve shut-off package to manage code preemption and
	// shutdown signaling.
	vv := valve.New()
	baseCtx := vv.Context()
	configFile := flag.String("config", "", "path to config file")
	envPrefix := os.Getenv("ENV_PREFIX")
	if envPrefix == "" {
		envPrefix = "app"
	}
	flag.Parse()

	srv, err := pkg.NewServer(baseCtx, *configFile, envPrefix)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			fmt.Println("shutting down..")

			// first valv
			vv.Shutdown(20 * time.Second)

			// create context with timeout
			ctx, cancel := context.WithTimeout(baseCtx, 20*time.Second)
			defer cancel()

			// start http shutdown
			srv.Shutdown(ctx)

			// verify, in worst case call cancel via defer
			select {
			case <-time.After(21 * time.Second):
				fmt.Println("not all connections done")
			case <-ctx.Done():

			}
		}
	}()
	log.Println("http server is listening...")
	srv.ListenAndServe()
}
