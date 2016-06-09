package etcd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Register is the helper function to self-register service into Etcd/Consul server
// name - service name
// host - service host
// port - service port
// target - etcd dial address, for example: "http://127.0.0.1:2379;http://127.0.0.1:12379"
// interval - interval of self-register to etcd
// ttl - ttl of the register information
func Register(name string, host string, port int, target string, interval time.Duration, ttl int) error {
	// get endpoints for register dial address
	endpoints := strings.Split(target, ";")
	conf := etcd.Config{
		Endpoints: endpoints,
	}

	client, err := etcd.New(conf)
	if err != nil {
		return errors.New(fmt.Sprintf("wonaming: create etcd client error: ", err.Error()))
	}
	keyapi := etcd.NewKeysAPI(client)

	serviceID := fmt.Sprintf("%s-%s-%d", name, host, port)
	serviceKey := fmt.Sprintf("/%s/%s/%s", prefix, name, serviceID)
	hostKey := fmt.Sprintf("/%s/%s/%s/host", prefix, name, serviceID)
	portKey := fmt.Sprintf("/%s/%s/%s/port", prefix, name, serviceID)

	//de-register if meet signhup
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
		log.Println("wonaming: receive signal: ", <-ch)

		_, err := keyapi.Delete(context.Background(), serviceKey, &etcd.DeleteOptions{Recursive: true})
		if err != nil {
			log.Println("wonaming: deregister service error: ", err)
		} else {
			log.Println("wonaming: deregistered service from etcd server.")
		}
		os.Exit(0)
	}()

	go func() {
		// invoke self-register with ticker
		ticker := time.NewTicker(interval)

		// refresh set to true for not notifying the watcher
		setopt := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second, Refresh: true}
		for {
			<-ticker.C
			if _, err := keyapi.Set(context.Background(), hostKey, "", setopt); err != nil {
				log.Println("wonaming: update ttl of host key error: ", err.Error())
			}
			if _, err := keyapi.Set(context.Background(), portKey, "", setopt); err != nil {
				log.Println("wonaming: update ttl of port key error: ", err.Error())
			}
		}
	}()

	// initial register
	setopt := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second}
	if _, err := keyapi.Set(context.Background(), hostKey, host, setopt); err != nil {
		return errors.New(fmt.Sprintf("wonaming: initial register service '%s' host to etcd error: %s", name, err.Error()))
	}
	if _, err := keyapi.Set(context.Background(), portKey, fmt.Sprintf("%d", port), setopt); err != nil {
		return errors.New(fmt.Sprintf("wonaming: initial register service '%s' port to etcd error: %s", name, err.Error()))
	}

	return nil
}
