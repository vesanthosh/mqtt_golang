package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/shirou/gopsutil/cpu"
)

type Object struct {
	CANConfig CANConfigArray
	CANData   CANDataArray
}

type CANConfigArray struct {
	DeviceId  string `json:"deviceId"`
	CANConfig []struct {
		CANId        string `json:"canId"`
		CANLabel     string `json:"canLabel"`
		TimeInterval int    `json:"timeInterval"`
	} `json:"canConfig"`
}

type CANDataArray struct {
	DeviceId  string `json:"deviceId"`
	CANConfig []struct {
		CANId    string `json:"canId"`
		CANLabel string `json:"canLabel"`
		CANData  string `json:"canData"`
	} `json:"canConfig"`
}

type CANDeviceInfoPayload struct {
	DeviceId string `json:"deviceId"`
}

func main() {
	localObject := &Object{}
	APIURL, err := url.Parse("http://localhost:3001/api/")
	if err != nil {
		fmt.Errorf("Error while parsing APIURL: %s", err)
	}
	deviceId := "ddeb27fb-d9a0-4624-be4d-4615062daed4"
	localObject.getCANInfo(APIURL, deviceId)
}

func (localStruct *Object) getCANInfo(URLString *url.URL, deviceId string) {
	topic := deviceId + "/canData"
	client := connect()

	timer := time.NewTicker(1 * time.Second)
	for t := range timer.C {
		localStruct.CANConfig, localStruct.CANData, _ = localStruct.getDeviceCANConfigs(URLString, deviceId)

		cpu, _ := cpu.Percent(0, true)

		for index, _ := range localStruct.CANConfig.CANConfig {
			// Temp concept
			canID, _ := strconv.ParseInt(localStruct.CANConfig.CANConfig[index].CANId, 10, 0)
			localStruct.CANData.CANConfig[index].CANData = fmt.Sprintf("%f", cpu[canID])
		}
		fmt.Println(t)

		jsonPayload, err := json.Marshal(localStruct.CANData)
		if err != nil {
			fmt.Errorf("Error while converting to json format: %s", err)
		}

		readableJSONData := fmt.Sprintf("%s", jsonPayload)

		client.Publish(topic, 1, true, readableJSONData)
		fmt.Printf("\n[%s] [%s] Published Data (MQTT): %s", t, topic, readableJSONData)
	}
}

//*****************mqtt connection******************
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
	opts.AddBroker("tcp://192.168.99.109:1883")
	opts.SetClientID("testClientID")
	return opts
}

//*****************mqtt connection******************

func (o *Object) getDeviceCANConfigs(URLString *url.URL, deviceId string) (resData CANConfigArray, resData1 CANDataArray, err error) {
	apiURL, err := url.Parse(URLString.String())
	if err != nil {
		fmt.Errorf("Unable to create URL for string: %s", apiURL)
		return resData, resData1, err
	}
	apiURL.Path = path.Join(apiURL.Path, "canconfig")
	apiURL.Path = path.Join(apiURL.Path, "all")

	deviceInfo := CANDeviceInfoPayload{}
	deviceInfo.DeviceId = deviceId
	canDevicePayload, err := json.Marshal(deviceInfo)
	if err != nil {
		return resData, resData1, err
	}
	// str := fmt.Sprintf("%s", canDevicePayload)
	// fmt.Println(str)
	// fmt.Println(apiURL)

	bytes, err := o.executeAPIQuery(apiURL, "GET", canDevicePayload)
	if err != nil {
		return resData, resData1, fmt.Errorf("Response Error - err:%+v", err)
	}
	err = json.Unmarshal(bytes, &resData)
	// some work around
	err = json.Unmarshal(bytes, &resData1)
	if err != nil {
		return resData, resData1, err
	}
	return
}

func (o *Object) executeAPIQuery(URLString *url.URL, reqType string, data []byte) (resData []byte, err error) {
	httpRequest, err := http.NewRequest(reqType, URLString.String(), bytes.NewBuffer(data))
	if err != nil {
		return resData, err
	}
	if len(data) > 0 {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	httpClient := &http.Client{}
	response, err := httpClient.Do(httpRequest)
	if err != nil {
		fmt.Errorf("Error httpclient do request: %+v", httpRequest)
		fmt.Errorf("Error httpclient do response: %+v", response)
		fmt.Errorf("Error httpclient do err: %+v", err)
		return resData, err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Errorf("Error httpclient do request: %+v", httpRequest)
		fmt.Errorf("Error httpclient do response: %+v", response)
		bytes, err2 := ioutil.ReadAll(response.Body)
		if err2 != nil {
			return resData, fmt.Errorf("error1:%+v, error2:%+v", err, err2)
		}
		fmt.Errorf("Error httpclient do response.body:%+v", string(bytes))
		return resData, fmt.Errorf("Unexpected response code:%+v", response.StatusCode, err)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return resData, errors.New("Unable to read the respone body")
	}
	return bytes, err
}
