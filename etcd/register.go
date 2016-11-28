package etcd

import (
	"fmt"
	"log"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

var client etcd.Client
var serviceKey string

// Register is the helper function to self-register service into Etcd/Consul server
// should call Unregister when pocess stop
// name - service name
// host - service host
// port - service port
// target - etcd dial address, for example: "http://127.0.0.1:2379;http://127.0.0.1:12379"
// interval - interval of self-register to etcd
// ttl - ttl of the register information
func Register(name string, host string, port int, target string, interval time.Duration, ttl int) error {
	// get endpoints for register dial address
	endpoints := strings.Split(target, ",")
	conf := etcd.Config{
		Endpoints: endpoints,
	}

	var err error
	client, err = etcd.New(conf)
	if err != nil {
		return fmt.Errorf("wonaming: create etcd client error: %v", err)
	}
	keyapi := etcd.NewKeysAPI(client)

	serviceID := fmt.Sprintf("%s-%s-%d", name, host, port)
	serviceKey = fmt.Sprintf("/%s/%s/%s", Prefix, name, serviceID)
	hostKey := fmt.Sprintf("/%s/%s/%s/host", Prefix, name, serviceID)
	portKey := fmt.Sprintf("/%s/%s/%s/port", Prefix, name, serviceID)

	go func() {
		// invoke self-register with ticker
		ticker := time.NewTicker(interval)

		// should get first, if not exist, set it
		for {
			<-ticker.C
			_, err := keyapi.Get(context.Background(), serviceKey, &etcd.GetOptions{Recursive: true})
			if err != nil {
				if _, err := keyapi.Set(context.Background(), hostKey, host, nil); err != nil {
					log.Printf("wonaming: re-register service '%s' host to etcd error: %s\n", name, err.Error())
				}
				if _, err := keyapi.Set(context.Background(), portKey, fmt.Sprintf("%d", port), nil); err != nil {
					log.Printf("wonaming: re-register service '%s' port to etcd error: %s\n", name, err.Error())
				}
				setopt := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second, PrevExist: etcd.PrevExist, Dir: true}
				if _, err := keyapi.Set(context.Background(), serviceKey, "", setopt); err != nil {
					log.Printf("wonaming: set service '%s' ttl to etcd error: %s\n", name, err.Error())
				}
			} else {
				// refresh set to true for not notifying the watcher
				setopt := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second, PrevExist: etcd.PrevExist, Dir: true, Refresh: true}
				if _, err := keyapi.Set(context.Background(), serviceKey, "", setopt); err != nil {
					log.Printf("wonaming: set service '%s' ttl to etcd error: %s\n", name, err.Error())
				}
			}
		}
	}()

	// initial register
	if _, err := keyapi.Set(context.Background(), hostKey, host, nil); err != nil {
		return fmt.Errorf("wonaming: initial register service '%s' host to etcd error: %s", name, err.Error())
	}
	if _, err := keyapi.Set(context.Background(), portKey, fmt.Sprintf("%d", port), nil); err != nil {
		return fmt.Errorf("wonaming: initial register service '%s' port to etcd error: %s", name, err.Error())
	}
	setopt := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second, PrevExist: etcd.PrevExist, Dir: true}
	if _, err := keyapi.Set(context.Background(), serviceKey, "", setopt); err != nil {
		return fmt.Errorf("wonaming: set service '%s' ttl to etcd error: %s", name, err.Error())
	}

	return nil
}

// Unregister delete service from etcd
func Unregister() error {
	keyapi := etcd.NewKeysAPI(client)
	_, err := keyapi.Delete(context.Background(), serviceKey, &etcd.DeleteOptions{Recursive: true})
	if err != nil {
		log.Println("wonaming: deregister service error: ", err.Error())
	} else {
		log.Println("wonaming: deregistered service from etcd server.")
	}
	return err
}
