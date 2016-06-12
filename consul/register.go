package consul

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// Register is the helper function to self-register service into Etcd/Consul server
// name - service name
// host - service host
// port - service port
// target - consul dial address, for example: "127.0.0.1:8500"
// interval - interval of self-register to etcd
// ttl - ttl of the register information
func Register(name string, host string, port int, target string, interval time.Duration, ttl int) error {
	conf := &consul.Config{Scheme: "http", Address: target}
	client, err := consul.NewClient(conf)
	if err != nil {
		return errors.New(fmt.Sprintf("wonaming: create consul client error: %s", err))
	}

	serviceID := fmt.Sprintf("%s-%s-%d", name, host, port)

	//de-register if meet signhup
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
		log.Println("wonaming: receive signal: ", <-ch)

		err := client.Agent().ServiceDeregister(serviceID)
		if err != nil {
			log.Println("wonaming: deregister service error: ", err.Error())
		} else {
			log.Println("wonaming: deregistered service from consul server.")
		}

		err = client.Agent().CheckDeregister(serviceID)
		if err != nil {
			log.Println("wonaming: deregister check error: ", err.Error())
		}

		os.Exit(0)
	}()

	// routine to update ttl
	go func() {
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C
			err = client.Agent().UpdateTTL(serviceID, "", "passing")
			if err != nil {
				log.Println("wonaming: update ttl of service error: ", err.Error())
			}
		}
	}()

	// initial register service
	regis := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: host,
		Port:    port,
	}
	err = client.Agent().ServiceRegister(regis)
	if err != nil {
		return errors.New(fmt.Sprintf("wonaming: initial register service '%s' host to consul error: %s", name, err.Error()))
	}

	// initial register service check
	check := consul.AgentServiceCheck{TTL: fmt.Sprintf("%ds", ttl), Status: "passing"}
	err = client.Agent().CheckRegister(&consul.AgentCheckRegistration{ID: serviceID, Name: name, ServiceID: serviceID, AgentServiceCheck: check})
	if err != nil {
		return errors.New(fmt.Sprintf("wonaming: initial register service check to consul error: %s", err.Error()))
	}

	return nil
}
