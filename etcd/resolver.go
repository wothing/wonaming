package etcd

import (
	"errors"
	"strings"

	etcd "github.com/coreos/etcd/client"
	"google.golang.org/grpc/naming"
)

type EtcdResolver struct {
	ServiceName string  // service name to resolve
}

func NewResolver(serviceName string) *EtcdResolver {
	return &EtcdResolver{ServiceName: serviceName}
}

func (er *EtcdResolver) Resolve(target string) (naming.Watcher, error) {
	if er.ServiceName == "" {
		return nil, errors.New("no service name provided")
	}

	endpoints := strings.Split(target, ";")
	conf := etcd.Config{
		Endpoints: endpoints,
	}

	client, err := etcd.New(conf)
	if err != nil {
		return nil, err
	}

	watcher := &EtcdWatcher{
		er: er,
		ec: &client,
	}
	return watcher, nil
}
