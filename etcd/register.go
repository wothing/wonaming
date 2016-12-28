/**
 * Copyright 2015-2016, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2016/12/19 17:40.
 */

package etcd

import (
	"fmt"
	"log"
	"strings"
	"time"

	etcd2 "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Prefix should start and end with no slash
var Prefix = "wonaming"
var client etcd2.Client
var serviceKey string

var stopSignal = make(chan bool, 1)

// Register
func Register(name string, host string, port int, target string, interval time.Duration, ttl int) error {
	serviceValue := fmt.Sprintf("%s:%d", host, port)
	serviceKey = fmt.Sprintf("/%s/%s/%s", Prefix, name, serviceValue)

	// get endpoints for register dial address
	var err error
	client, err = etcd2.New(etcd2.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		return fmt.Errorf("wonaming: create etcd2 client failed: %v", err)
	}

	keysAPI := etcd2.NewKeysAPI(client)
	go func() {
		// invoke self-register with ticker
		ticker := time.NewTicker(interval)
		setOptions := &etcd2.SetOptions{TTL: time.Second * time.Duration(ttl), Refresh: true, PrevExist: etcd2.PrevExist}
		for {
			// should get first, if not exist, set it
			_, err := keysAPI.Get(context.Background(), serviceKey, &etcd2.GetOptions{Recursive: true})
			if err != nil {
				if etcd2.IsKeyNotFound(err) {
					if _, err := keysAPI.Set(context.Background(), serviceKey, serviceValue, &etcd2.SetOptions{TTL: time.Second * time.Duration(ttl)}); err != nil {
						log.Printf("wonaming: set service '%s' with ttl to etcd2 failed: %s", name, err.Error())
					}
				} else {
					log.Printf("wonaming: service '%s' connect to etcd2 failed: %s", name, err.Error())
				}
			} else {
				// refresh set to true for not notifying the watcher
				if _, err := keysAPI.Set(context.Background(), serviceKey, "", setOptions); err != nil {
					log.Printf("wonaming: refresh service '%s' with ttl to etcd2 failed: %s", name, err.Error())
				}
			}
			select {
			case <-stopSignal:
				return
			case <-ticker.C:
			}
		}
	}()

	return nil
}

// UnRegister delete registered service from etcd
func UnRegister() error {
	stopSignal <- true
	stopSignal = make(chan bool, 1) // just a hack to avoid multi UnRegister deadlock
	_, err := etcd2.NewKeysAPI(client).Delete(context.Background(), serviceKey, &etcd2.DeleteOptions{Recursive: true})
	if err != nil {
		log.Printf("wonaming: deregister '%s' failed: %s", serviceKey, err.Error())
	} else {
		log.Printf("wonaming: deregister '%s' ok.", serviceKey)
	}
	return err
}
