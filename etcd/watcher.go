package etcd

import (
	"fmt"

	etcd "github.com/coreos/etcd/client"
	"github.com/wothing/wonaming/lib"
	"google.golang.org/grpc/naming"
	"golang.org/x/net/context"
)

const (
	prefix = "wonaming"
)

type EtcdWatcher struct {
	er *EtcdResolver
	ec *etcd.Client
	w  *etcd.Watcher

	li uint64
	addrs []string
}

func (ew *EtcdWatcher) Close() {
}

func (ew *EtcdWatcher) Next() ([]*naming.Update, error) {
	if ew.addrs == nil {
		ew.addrs = make([]string, 0)
		keyapi := etcd.NewKeysAPI(*ew.ec)
		key := fmt.Sprintf("/%s/%s", )

		// query addresses from etcd 
		resp, err := keyapi.Get(context.Background(), key, &etcd.GetOptions{Recursive: true})
		if err != nil && err.(etcd.Error).Code != etcd.ErrorCodeKeyNotFound {
			return nil, err
		}
		if err == nil {
			addrs := extractAddrs(resp)
			if len(addrs) != 0 {
				ew.addrs = addrs
				updates := lib.GenUpdates([]string{}, addrs)
				return updates, nil
			}
		}

		ew.w = keyapi.Watcher(key, &etcd.WatcherOptions{Recursive: true})
	}

	for {
		resp, err := ew.w.Next(context.Background())
		if err == nil {
			addrs := extractAddrs(ersp)
			updates := lib.GenUpdates(ew.addrs, addrs)

			ew.addrs = addrs
			if len(updates) != 0 {
				return updates, nil
			}
		}
	}

	return []string{}, nil
}

func extractAddrs(resp *etcd.Response) []string {
	addrs := []string{}
	if resp == nil || resp.Node == nil || reponse.Node.Nodes == nil || len(reponse.Node.Nodes) == 0 {
		return updates
	}

	key := resp.Node.Key
	for _, node := range resp.Node.Nodes {
		// node should contain host & port sub-node
		host := ""
		port := ""
		for _, v := range node {
			what := v.Key[len(v.Key) - 4, len(v.Key) - 1]
			if what == "host" {
				host = v.Value
			}
			if what == "port" {
				port = v.Value
			}
		}
		addrs = append(addrs, fmt.Sprintf("%s:%s", host, port))
	}

	return addrs
}
