package main

import (
	"bufio"
	"cs.ubc.ca/cpsc416/p1/raftkv"
	"cs.ubc.ca/cpsc416/p1/util"
	"github.com/DistributedClocks/tracing"
	"log"
	"os"
	"strings"
)

type ClientConfig struct {
	ClientID          string
	LocalServerIPPort string
	ServerIPPortList  []string
	ChCapacity        int
	TracingServerAddr string
	Secret            []byte
	TracingIdentity   string
}

var (
	resetColour   = "\033[0m"  // default text colour
	successColour = "\033[32m" // green text
)

func main() {
	var config ClientConfig
	err := util.ReadJSONConfig("config/client_config.json", &config)
	util.CheckErr(err, "Error reading client config: %v\n", err)

	tracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})

	client := raftkv.NewKVS()
	notifCh, err := client.Start(tracer, config.ClientID, config.LocalServerIPPort, config.ServerIPPortList, config.ChCapacity)
	util.CheckErr(err, "Error reading client config: %v\n", err)

	if len(os.Args) == 2 && os.Args[1] == "-i" {
		runInteractiveClient(client, notifCh)
	} else {
		runTestScript(client, notifCh)
	}
}

func runTestScript(client *raftkv.KVS, notifCh raftkv.NotifyChannel) {
	// Put a key-value pair
	err := client.Put("key2", "value2")
	util.CheckErr(err, "Error putting value %v, opId: %v\b", err)

	// Get a key's value
	err = client.Get("key1")
	util.CheckErr(err, "Error getting value %v, opId: %v\b", err)

	// Sequence of interleaved gets and puts
	err = client.Put("key1", "test1")
	util.CheckErr(err, "Error putting value %v, opId: %v\b", err)
	err = client.Get("key1")
	util.CheckErr(err, "Error getting value %v, opId: %v\b", err)
	err = client.Put("key1", "test2")
	util.CheckErr(err, "Error putting value %v, opId: %v\b", err)
	err = client.Get("key1")
	util.CheckErr(err, "Error getting value %v, opId: %v\b", err)
	err = client.Get("key1")
	util.CheckErr(err, "Error getting value %v, opId: %v\b", err)
	err = client.Put("key1", "test3")
	util.CheckErr(err, "Error putting value %v, opId: %v\b", err)
	err = client.Get("key1")
	util.CheckErr(err, "Error getting value %v, opId: %v\b", err)

	for i := 0; i < 9; i++ {
		result := <-notifCh
		log.Printf("%s%v%s\n", successColour, result, resetColour)
	}
	client.Stop()
}

// Run client in an interactive command line
// e.g. 'put k1 v1' or 'get k1'
func runInteractiveClient(client *raftkv.KVS, notifyCh raftkv.NotifyChannel) {
	defer func() {
		client.Stop()
		log.Println("Session terminated")
	}()

	go func() {
		// Print results as they return from KVS
		for result := range notifyCh {
			log.Printf("%s%v%s\n", successColour, result, resetColour)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	// Parse and execute operations input by user
	for {
		scanner.Scan()
		text := scanner.Text()
		args := strings.Fields(text)
		if len(args) == 0 {
			// Terminate session on empty line
			break
		}

		op := args[0]
		if len(args) == 2 && op == "get" {
			key := args[1]
			err := client.Get(key)
			util.CheckErr(err, "Error getting value at key %s", key)
			continue
		}
		if len(args) == 3 && op == "put" {
			key := args[1]
			value := args[2]
			err := client.Put(key, value)
			util.CheckErr(err, "Error putting value %s to key %s", value, key)
			continue
		}
		log.Println("Invalid command")
	}
}
