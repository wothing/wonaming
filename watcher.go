package wonaming

import (
	"errors"
	"fmt"
	"time"

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
	if cw.addrs == nil {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				addrs, li, err := cw.queryConsul(nil)
				if err != nil {
					return nil, err
				}
				if len(addrs) != 0 {
					cw.addrs = addrs
					cw.li = li
					updates := genUpdates([]string{}, addrs)
					return updates, nil
				}
			}
		}
	}

	addrs, li, err := cw.queryConsul(&consul.QueryOptions{WaitIndex: cw.li})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("consul query error: %v", err))
	}

	updates := genUpdates(cw.addrs, addrs)

	// update addrs & last index
	cw.addrs = addrs
	cw.li = li

	return updates, nil
}

func (cw *ConsulWatcher) queryConsul(q *consul.QueryOptions) ([]string, uint64, error) {
	cs, meta, err := cw.cc.Health().Service(cw.cr.ServiceName, "", true, q)
	if err != nil {
		return nil, 0, errors.New(fmt.Sprintf("consul query error: %v", err))
	}

	addrs := make([]string, 0)
	for _, s := range cs {
		// addr should like: 127.0.0.1:8001
		addrs = append(addrs, fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port))
	}

	return addrs, meta.LastIndex, nil
}

func genUpdates(a, b []string) []*naming.Update {
	updates := []*naming.Update{}

	deleted := diff(a, b)
	for _, addr := range deleted {
		update := &naming.Update{Op: naming.Delete, Addr: addr}
		updates = append(updates, update)
	}

	added := diff(b, a)
	for _, addr := range added {
		update := &naming.Update{Op: naming.Add, Addr: addr}
		updates = append(updates, update)
	}
	return updates
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
