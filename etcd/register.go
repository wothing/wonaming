package etcd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func Register(sName string, sPort int, etcdAddr string, interval time.Duration, ttl int) error {
	sAddr := "127.0.0.1"
	env := os.Getenv("HOSTNAME")
	if env != "" {
		sAddr = env
	}

	endpoints := strings.Split(etcdAddr, ";")
	log.Printf("endpoints: %v\n", endpoints)
	conf := etcd.Config{
		Endpoints: endpoints,
	}

	client, err := etcd.New(conf)
	if err != nil {
		return err
	}

	serviceID := fmt.Sprintf("%s-%s-%d", sName, sAddr, sPort)
	serviceKey := fmt.Sprintf("/%s/%s/%s", prefix, sName, serviceID)
	hostKey := fmt.Sprintf("/%s/%s/%s/host", prefix, sName, serviceID)
	portKey := fmt.Sprintf("/%s/%s/%s/port", prefix, sName, serviceID)

	keyapi := etcd.NewKeysAPI(client)

	//de-register if meet signhup
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		log.Println("rec signal:", <-ch)

		_, err := keyapi.Delete(context.Background(), serviceKey, &etcd.DeleteOptions{Recursive: true})
		if err != nil {
			log.Println("deregister service error: ", err)
		}
		os.Exit(0)
	}()

	go func() {
		ticker := time.NewTicker(interval)
		seto := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second, Refresh: true}
		for {
			<-ticker.C
			if _, err := keyapi.Set(context.Background(), hostKey, "", seto); err != nil {
				log.Println("update ttl error: ", err)
			}
			if _, err := keyapi.Set(context.Background(), portKey, "", seto); err != nil {
				log.Println("update ttl error: ", err)
			}
		}
	}()

	seto := &etcd.SetOptions{TTL: time.Duration(ttl) * time.Second}
	if _, err := keyapi.Set(context.Background(), hostKey, sAddr, seto); err != nil {
		return err
	}
	if _, err := keyapi.Set(context.Background(), portKey, fmt.Sprintf("%d", sPort), seto); err != nil {
		return err
	}

	return nil
}
