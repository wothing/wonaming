package consul 

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	consul "github.com/hashicorp/consul/api"
)

// Register helps register consul
func Register(sName string, sPort int, c string, interval string) error {
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

	check := &consul.AgentServiceCheck{Interval: interval, Script: fmt.Sprintf(`curl http://%s:%d > /dev/null 2>&1`, sAddr, sPort)}
	serviceID := fmt.Sprintf("%s-%s-%d", sName, sAddr, sPort)
	regis := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    sName,
		Address: sAddr,
		Port:    sPort,
		Checks:  consul.AgentServiceChecks{check},
	}

	//de-register if meet signhup
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		<-ch
		log.Println("rec signal, deregister service, result:", client.Agent().ServiceDeregister(serviceID))
		os.Exit(0)
	}()

	return client.Agent().ServiceRegister(regis)
}
