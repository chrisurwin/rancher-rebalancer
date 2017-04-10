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

//Global Variable
var cattleURL = os.Getenv("CATTLE_URL")
var cattleAccessKey = os.Getenv("CATTLE_ACCESS_KEY")
var cattleSecretKey = os.Getenv("CATTLE_SECRET_KEY")
var opts = &client.ClientOpts{
	Url:       cattleURL,
	AccessKey: cattleAccessKey,
	SecretKey: cattleSecretKey,
}

var projectID = ""
var rancherEnv = ""

type serviceDef struct {
	name        string
	id          string
	instanceIds []string
}

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
	for {
		var returnCode = 0
		returnCode = rebalance()
		if returnCode == 1 {
			time.Sleep(60 * time.Second)
		} else {
			time.Sleep(5 * time.Minute)
		}
	}
}

//Function to rebalance containers between hosts
func rebalance() int {
	var hostList = hostIdList()
	var serviceInstanceID = serviceIDList()
	for service := range serviceInstanceID {
		logrus.Info("Currently processing service: " + service)
		serviceHosts(serviceInstanceID[service], hostList)
	}
	logrus.Info("Finished current processing")
	return 1
}
