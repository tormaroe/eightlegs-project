package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tormaroe/eightlegs-project/pinkfoot/api"
	"github.com/tormaroe/eightlegs-project/pinkfoot/config"
	"github.com/tormaroe/eightlegs-project/pinkfoot/queue"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: pinkfoot <config.toml>")
		os.Exit(1)
	}

	config, err := config.Load(os.Args[1])
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(2)
	}

	q, err := queue.Init(config)
	if err != nil {
		fmt.Println("Error initializing queue:", err)
		os.Exit(3)
	}

	apiHandler := api.Handler{
		Queue: q,
	}
	http.Handle("/", &apiHandler)

	log.Printf("Listening to port %d\n", config.API.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.API.Port), nil)
}
