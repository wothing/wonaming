/**
 * Copyright 2015-2016, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2016/12/19 17:40.
 */

package etcd

import (
	"errors"
	"fmt"
	"strings"

	etcd2 "github.com/coreos/etcd/client"
	"google.golang.org/grpc/naming"
)

// resolver is the implementaion of grpc.naming.Resolver
type resolver struct {
	serviceName string // service name to resolve
}

// NewResolver return resolver with service name
func NewResolver(serviceName string) *resolver {
	return &resolver{serviceName: serviceName}
}

// Resolve to resolve the service from etcd, target is the dial address of etcd
// target example: "http://127.0.0.1:2379,http://127.0.0.1:12379,http://127.0.0.1:22379"
func (re *resolver) Resolve(target string) (naming.Watcher, error) {
	if re.serviceName == "" {
		return nil, errors.New("wonaming: no service name provided")
	}

	// generate etcd client
	client, err := etcd2.New(etcd2.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("wonaming: creat etcd2 client failed: %s", err.Error())
	}

	// Return watcher
	return &watcher{re: re, api: etcd2.NewKeysAPI(client)}, nil
}
