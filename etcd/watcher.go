/**
 * Copyright 2015-2016, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2016/12/19 17:40.
 */

package etcd

import (
	"fmt"
	"strings"

	etcd2 "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
)

// watcher is the implementaion of grpc.naming.Watcher
type watcher struct {
	re            *resolver // re: Etcd Resolver
	api           etcd2.KeysAPI
	isInitialized bool
}

// Close do nothing
func (w *watcher) Close() {
}

// Next to return the updates
func (w *watcher) Next() ([]*naming.Update, error) {
	// dir is the etcd dir/value dir to watch
	dir := fmt.Sprintf("/%s/%s/", Prefix, w.re.serviceName)

	// check if is initialized
	if !w.isInitialized {
		// query addresses from etcd
		resp, err := w.api.Get(context.Background(), dir, &etcd2.GetOptions{Recursive: true})
		w.isInitialized = true
		if err == nil {
			addrs := extractAddrs(resp)
			//if not empty, return the updates or watcher new dir
			if l := len(addrs); l != 0 {
				updates := make([]*naming.Update, l)
				for i := range addrs {
					updates[i] = &naming.Update{Op: naming.Add, Addr: addrs[i]}
				}
				return updates, nil
			}
		}
	}

	// generate etcd Watcher
	etcdWatcher := w.api.Watcher(dir, &etcd2.WatcherOptions{Recursive: true})
	for {
		resp, err := etcdWatcher.Next(context.Background())
		if err == nil {
			switch resp.Action {
			case "set":
				return []*naming.Update{{Op: naming.Add, Addr: resp.Node.Value}}, nil
			case "delete", "expire":
				return []*naming.Update{{Op: naming.Delete, Addr: strings.TrimPrefix(resp.Node.Key, dir)}}, nil // not using PrevNode because it may nil
			}
		}
	}
}

func extractAddrs(resp *etcd2.Response) []string {
	addrs := []string{}

	if resp == nil || resp.Node == nil {
		return addrs
	}

	for i := range resp.Node.Nodes {
		if v := resp.Node.Nodes[i].Value; v != "" {
			addrs = append(addrs, v)
		}
	}

	return addrs
}
