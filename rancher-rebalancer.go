package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
)

var (
	cattleURL       = os.Getenv("CATTLE_URL")
	cattleAccessKey = os.Getenv("CATTLE_ACCESS_KEY")
	cattleSecretKey = os.Getenv("CATTLE_SECRET_KEY")
	opts            = &client.ClientOpts{
		Url:       cattleURL,
		AccessKey: cattleAccessKey,
		SecretKey: cattleSecretKey,
	}
	projectID  = ""
	rancherEnv = ""
	mode       = os.Getenv("MODE")
	opt        = os.Getenv("OPT")
)

var rancherServices = map[string]string{
	"healthcheck":     "",
	"ipsec":           "",
	"network-manager": "",
	"scheduler":       "",
	"metadata":        "",
	"rancher-agent1":  ""}

func main() {
	if len(cattleURL) == 0 {
		logrus.Fatalf("CATTLE_URL is not set")
	}

	if len(cattleAccessKey) == 0 {
		logrus.Fatalf("CATTLE_ACCESS_KEY is not set")
	}

	if len(cattleSecretKey) == 0 {
		logrus.Fatalf("CATTLE_SECRET_KEY is not set")
	}
	//Get the Environment
	if os.Getenv("ENVIRONMENT") == "" {
		resp, err := http.Get("http://rancher-metadata/latest/self/stack/environment_name")
		if err != nil {
			fmt.Println("Rancher Metadata not available")
		} else {
			defer resp.Body.Close()
			respOutput, _ := ioutil.ReadAll(resp.Body)
			rancherEnv = string(respOutput)
			fmt.Println("Rancher environment set to: " + rancherEnv)
		}
	} else {
		rancherEnv = os.Getenv("ENVIRONMENT")
	}
	if rancherEnv == "" {
		logrus.Fatal("Rancher Environment not found")
	}

	logrus.Info("Starting Rancher Rebalancer")
	projectID = getProjectID(rancherEnv)
	go startHealthcheck()
	logrus.Info("Operating Mode: " + mode + " Opt mode: " + opt)
	for {
		var returnCode = 0
		returnCode = rebalance()
		time.Sleep(time.Duration(returnCode) * time.Minute)
	}
}

//Function to rebalance containers between nodes
func rebalance() int {

	var balanced = false
	if mode != "AGGRESSIVE" {
		balanced = evenLoad()
	}

	if !balanced {
		var hostList = hostIDList()
		var serviceInstanceID = serviceIDList()

		for service := range serviceInstanceID {
			logrus.Info("Currently processing service: " + service)
			serviceHosts(serviceInstanceID[service], hostList)
		}

	} else {

		logrus.Info("Server load balanced, AGGRESSIVE mode would need to be used to enforce container balancing")
		return 1

	}

	return 1
}
