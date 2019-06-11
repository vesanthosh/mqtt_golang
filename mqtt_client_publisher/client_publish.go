package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type JSONData struct {
	Platform  string `json:"platform"`
	Version   string `json:"version"`
	Processes struct {
		Running int `json:"running"`
		Blocked int `json:"blocked"`
	} `json:"processes"`
	MemoryUsage struct {
		TolalSpace uint64  `json:"totalSpace"`
		FreeSpace  uint64  `json:"freeSpace"`
		MemoryUsed float64 `json:"memoryUsed"`
	} `json:"memoryUsage"`
	Time string `json:"time"`
}

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

		platform, version, _, _ := host.PlatformInformation()
		misc, _ := load.Misc()
		v, _ := mem.VirtualMemory()

		jsonBody := JSONData{}
		jsonBody.Platform = platform
		jsonBody.Version = version
		jsonBody.Processes.Running = misc.ProcsRunning
		jsonBody.Processes.Blocked = misc.ProcsBlocked
		jsonBody.MemoryUsage.TolalSpace = v.Total
		jsonBody.MemoryUsage.FreeSpace = v.Free
		jsonBody.MemoryUsage.MemoryUsed = v.UsedPercent
		jsonBody.Time = t.String()

		jsonPayload, err := json.Marshal(jsonBody)
		if err != nil {
			fmt.Errorf("Error while converting to json format: %s", err)
		}
		readableJSONData := fmt.Sprintf("%s", jsonPayload)

		client.Publish(topic, 1, true, readableJSONData)
		fmt.Println("Published Data (MQTT): ", readableJSONData)
	}
}
