package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
)

//Function to return the health state of an environemnt
func envHealth(e string) string {

	c, err := client.NewRancherClient(opts)
	if err != nil {
		logrus.Error("Error with client connection")
	}
	stacks, err := c.Project.List(nil)
	if err != nil {
		logrus.Error("Error with stack list")
	}

	for _, p := range stacks.Data {
		if p.Name == rancherEnv {
			return p.HealthState
		}
	}

	logrus.Error("Environment " + e + " not found")
	return "NotFound"
}

//Function to return the projectID of an environemnt
func getProjectID(e string) string {

	c, err := client.NewRancherClient(opts)
	if err != nil {
		logrus.Error("Error with client connection")
	}
	stacks, err := c.Project.List(nil)
	if err != nil {
		logrus.Error("Error with stack list")
	}

	for _, p := range stacks.Data {
		if p.Name == rancherEnv {
			logrus.Info("Environment projectid: " + p.Id)
			return p.Id
		}
	}

	logrus.Error("Environment " + e + " not found")
	return "NotFound"
}

//Function to evacuate a host
func evacuateHost(hostName string) bool {
	c, err := client.NewRancherClient(opts)
	if err != nil {
		logrus.Error("Error with evacuateHost client connection")
		return false
	}
	//Get a list of Hosts
	hosts, err := c.Host.List(nil)
	for _, h := range hosts.Data {
		if h.Hostname == hostName {
			c.Host.ActionEvacuate(&h)
		}
	}
	return true
}

//Function to return a list of hostids
func hostIdList() map[string]int {
	hostIds := map[string]int{}
	c, err := client.NewRancherClient(opts)
	if err != nil {
		logrus.Error("Error with hostID client connection")
	}

	//Get a list of Hosts in this environment

	hosts, err := c.Host.List(nil)
	for _, h := range hosts.Data {
		if h.AccountId == projectID && h.State == "active" {
			hostIds[h.Id] = 0
		}
	}
	return hostIds
}

//Function to return a list of services
func serviceIDList() map[string]serviceDef {
	var service = map[string]serviceDef{}
	c, err := client.NewRancherClient(opts)
	if err != nil {
		logrus.Error("Error with serviceID client connection")
	}

	//Get a list of Hosts in this environment

	services, err := c.Service.List(nil)
	for _, h := range services.Data {

		if h.AccountId == projectID {
			if _, exists := rancherServices[h.Name]; exists {
				//Ignore it as its a Rancher Service
			} else {
				var dat map[string]interface{}
				var affinity = false
				//Exclude the service is rebalance label is set to false
				labels, _ := json.Marshal(h.LaunchConfig.Labels)
				if err := json.Unmarshal(labels, &dat); err != nil {
					panic(err)
				}

				//Check for affinity rules
				for k, _ := range dat {
					if strings.Index(k, "scheduler.affinity") > 0 {
						affinity = true
					}
				}

				if affinity {
					logrus.Info("Not balancing " + h.Name + " due to affinity rules")
				} else {
					if val, ok := dat["rebalance"]; ok {
						if val == "false" {
							break
						}
					}
					var tempService serviceDef
					tempService.id = h.Id
					tempService.name = h.Name
					tempService.instanceIds = h.InstanceIds
					service[h.Name] = tempService
				}
			}
		}
	}
	return service
}

func serviceHosts(service serviceDef, hostList map[string]int) int {
	hosts := make(map[string]int)
	containers := make(map[string]string)
	for k2 := range hostList {
		hosts[k2] = 0
	}
	var instanceIDs = service.instanceIds
	for i := 0; i < len(instanceIDs); i++ {
		hostID := getContainerHost(instanceIDs[i])
		for host := range hosts {
			if host == hostID {
				hosts[host] = hosts[host] + 1
			}
		}
		containers[instanceIDs[i]] = hostID
	}

	var average = roundCount(len(hosts), len(instanceIDs))

	high := ""
	for host := range hosts {
		if hosts[host] > average {
			high = host
		}
	}

	if high != "" {
		//Need to delete a container from this host
		//first find a container
		for instance := range containers {
			if containers[instance] == high {

				cattleURLv2 := strings.Replace(cattleURL, "/v1", "/v2-beta", -1)
				var opts2 = &client.ClientOpts{
					Url:       cattleURLv2 + "/projects/" + projectID + "/schemas",
					AccessKey: cattleAccessKey,
					SecretKey: cattleSecretKey,
				}
				cc, err := client.NewRancherClient(opts2)
				if err != nil {
					logrus.Error("Error with container client connection")
				}

				//Delete the container
				containerToDelete, err := cc.Container.ById(instance)
				cc.Container.Delete(containerToDelete)
				if err != nil {
					logrus.Error(err)
				}

				//Wait for 10 seconds to allow for allocations service to allocate new server
				time.Sleep(10 * time.Second)

				break
			}
		}
	}

	return 0
}

func getContainerHost(id string) string {
	c, err := client.NewRancherClient(opts)
	if err != nil {
		logrus.Error("Error with client connection")
	}

	services, err := c.Container.ById(id)
	return services.HostId
}
