package wonaming

import (
	"errors"
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

type ConsulWatcher struct {
	cr *ConsulResolver
	cc *consul.Client

	// LastIndex to watch consul
	li uint64

	// before check: every value shoud be 1
	// after check: 1 - deleted  2 - nothing  3 - new added
	addrs []string
}

func (cw *ConsulWatcher) Close() {
}

func (cw *ConsulWatcher) Next() ([]*naming.Update, error) {
	var q *consul.QueryOptions = nil
	if cw.addrs != nil {
		q = &consul.QueryOptions{WaitIndex: cw.li}
	}

	updates := []*naming.Update{}

	// cs, meta, err := cw.cc.Catalog().Service(cw.cr.ServiceName, "", queryo)
	cs, meta, err := cw.cc.Health().Service(cw.cr.ServiceName, "", true, q)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("consul query error: %v", err))
	}

	addrs := make([]string, 0)
	for _, s := range cs {
		// addr should like: 127.0.0.1:8001
		addrs = append(addrs, fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port))
	}

	deleted := diff(cw.addrs, addrs)
	for _, addr := range deleted {
		update := &naming.Update{Op: naming.Delete, Addr: addr}
		updates = append(updates, update)
	}

	added := diff(addrs, cw.addrs)
	for _, addr := range added {
		update := &naming.Update{Op: naming.Add, Addr: addr}
		updates = append(updates, update)
	}

	// update addrs & last index
	cw.addrs = addrs
	cw.li = meta.LastIndex

	return updates, nil
}

// diff(a, b) = a - a(n)b
func diff(a, b []string) []string {
	d := make([]string, 0)
	for _, va := range a {
		found := false
		for _, vb := range b {
			if va == vb {
				found = true
				break
			}
		}

		if !found {
			d = append(d, va)
		}
	}
	return d
}
