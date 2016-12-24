package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/zenja/gsr/server"
)

type GSRConf struct {
	Port       int    `json:"port"`
	APIKey     string `json:"apiKey"`
	EngineID   string `json:"engineID"`
	TimeoutStr string `json:"timeout"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify config file.")
	}
	confFn := os.Args[1]

	f, err := os.Open(confFn)
	if err != nil {
		log.Fatalf("Error opening conf file %s: %s", confFn, err)
	}

	decoder := json.NewDecoder(f)
	var conf GSRConf
	if err := decoder.Decode(&conf); err != nil {
		log.Fatalf("Failed to decode config file %s: %s", confFn, err)
	}
	if len(conf.APIKey) == 0 {
		log.Fatal("API key is empty")
	}
	if len(conf.EngineID) == 0 {
		log.Fatal("Engine ID is empty")
	}
	f.Close()

	log.Printf("Starting server at port %d", conf.Port)
	timeout, err := time.ParseDuration(conf.TimeoutStr)
	if err != nil {
		log.Fatalf("Failed to parse timeout duration string \"%s\": %s", conf.TimeoutStr, err)
	}
	server.StartServer(conf.Port, conf.APIKey, conf.EngineID, timeout)
}
