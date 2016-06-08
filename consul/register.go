package consul

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// Register helps register consul
func Register(sName string, sPort int, c string, interval time.Duration, ttl string) error {
	sAddr := "127.0.0.1"
	env := os.Getenv("HOSTNAME")
	if env != "" {
		sAddr = env
	}

	conf := &consul.Config{Scheme: "http", Address: c}
	client, err := consul.NewClient(conf)
	if err != nil {
		return err
	}

	serviceID := fmt.Sprintf("%s-%s-%d", sName, sAddr, sPort)
	regis := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    sName,
		Address: sAddr,
		Port:    sPort,
	}

	//de-register if meet signhup
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		log.Println("rec signal:", <-ch)

		err = client.Agent().ServiceDeregister(serviceID)
		if err != nil {
			log.Println("deregister service, result:", err)
		}
		err = client.Agent().CheckDeregister(serviceID)
		if err != nil {
			log.Println("deregister check, result:", err)
		}

		os.Exit(0)
	}()

	go func() {
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C
			err = client.Agent().UpdateTTL(serviceID, "", "passing")
			if err != nil {
				log.Println("update ttl error", err)
			}
		}
	}()

	err = client.Agent().ServiceRegister(regis)
	if err != nil {
		return err
	}

	check := consul.AgentServiceCheck{TTL: ttl, Status: "passing"}
	return client.Agent().CheckRegister(&consul.AgentCheckRegistration{ID: serviceID, Name: sName, ServiceID: serviceID, AgentServiceCheck: check})
}
