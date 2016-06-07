package wonaming

import (
	"errors"

	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

type ConsulResolver struct {
	ServiceName string	//service name
}

func NewResolver(serviceName string) *ConsulResolver {
	return &ConsulResolver{ServiceName: serviceName}
}

func (cr *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	if cr.ServiceName == "" {
		return nil, errors.New("no service name provided")
	}

	conf := &consul.Config{
		Scheme: "http",
		Address: target,
	}
	client, err := consul.NewClient(conf)
	if err != nil {
		return nil, err
	}

	watcher := &ConsulWatcher{
		cr: cr,
		cc: client,
	}
	return watcher, nil
}
