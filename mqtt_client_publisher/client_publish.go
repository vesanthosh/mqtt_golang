package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func connect() mqtt.Client {
	opts := createClientOptions()
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://192.168.99.100:1883")
	opts.SetClientID("testClientID")
	return opts
}

func main() {
	topic := "test/data"

	client := connect()
	timer := time.NewTicker(1 * time.Second)
	for t := range timer.C {
		client.Publish(topic, 1, true, t.String())
		fmt.Println("Published Data (MQTT): ", t.String())
	}
}
