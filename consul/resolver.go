package consul

import (
	"errors"
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

// ConsulResolver is the implementaion of grpc.naming.Resolver
type ConsulResolver struct {
	ServiceName string //service name
}

// NewResolver return ConsulResolver with service name
func NewResolver(serviceName string) *ConsulResolver {
	return &ConsulResolver{ServiceName: serviceName}
}

// Resolve to resolve the service from consul, target is the dial address of consul
func (cr *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	if cr.ServiceName == "" {
		return nil, errors.New("wonaming: no service name provided")
	}

	// generate consul client, return if error
	conf := &consul.Config{
		Scheme:  "http",
		Address: target,
	}
	client, err := consul.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("wonaming: creat consul error: %v", err)
	}

	// return ConsulWatcher
	watcher := &ConsulWatcher{
		cr: cr,
		cc: client,
	}
	return watcher, nil
}
