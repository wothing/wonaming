package etcd

import (
	"fmt"

	etcd "github.com/coreos/etcd/client"
	"github.com/wothing/wonaming/lib"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
)

const (
	prefix = "wonaming"
)

type EtcdWatcher struct {
	er *EtcdResolver
	ec *etcd.Client
	w  etcd.Watcher

	addrs []string
}

func (ew *EtcdWatcher) Close() {
}

func (ew *EtcdWatcher) Next() ([]*naming.Update, error) {
	keyapi := etcd.NewKeysAPI(*ew.ec)
	key := fmt.Sprintf("/%s/%s", prefix, ew.er.ServiceName)

	if ew.addrs == nil {
		ew.addrs = make([]string, 0)

		// query addresses from etcd
		resp, err := keyapi.Get(context.Background(), key, &etcd.GetOptions{Recursive: true})
		if err != nil {
			etcderr, ok := err.(etcd.Error)
			if !ok || etcderr.Code != etcd.ErrorCodeKeyNotFound {
				return nil, err
			}
		}
		if err == nil {
			addrs := extractAddrs(resp)
			if len(addrs) != 0 {
				ew.addrs = addrs
				updates := lib.GenUpdates([]string{}, addrs)
				return updates, nil
			}
		}
	}

	ew.w = keyapi.Watcher(key, &etcd.WatcherOptions{Recursive: true})
	for {
		_, err := ew.w.Next(context.Background())
		if err == nil {
			// query addresses from etcd
			resp, _:= keyapi.Get(context.Background(), key, &etcd.GetOptions{Recursive: true})
			addrs := extractAddrs(resp)
			updates := lib.GenUpdates(ew.addrs, addrs)

			ew.addrs = addrs
			if len(updates) != 0 {
				return updates, nil
			}
		}
	}

	return []*naming.Update{}, nil
}

func extractAddrs(resp *etcd.Response) []string {
	addrs := []string{}
	if resp == nil || resp.Node == nil || resp.Node.Nodes == nil || len(resp.Node.Nodes) == 0 {
		return addrs
	}

	for _, node := range resp.Node.Nodes {
		// node should contain host & port sub-node
		host := ""
		port := ""
		for _, v := range node.Nodes {
			what := v.Key[len(v.Key)-4 : len(v.Key)]
			if what == "host" {
				host = v.Value
			}
			if what == "port" {
				port = v.Value
			}
		}
		if host != "" && port != "" {
			addrs = append(addrs, fmt.Sprintf("%s:%s", host, port))
		}
	}

	return addrs
}
