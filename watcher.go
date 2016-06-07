package wonaming

import (
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

type ConsulWatcher struct {
	cr *ConsulResolver
	cc *consul.Client
	li uint64

	// before check: every value shoud be 1
	// after check: 1 - deleted  2 - nothing  3 - new added
	addrs []string
}

func (cw *ConsulWatcher) Close() {
}

func (cw *ConsulWatcher) Next() ([]*naming.Update, error) {
	if cw.addrs == nil {
		cw.addrs = make([]string, 0)
	}

	updates := []*naming.Update{}

	t := time.NewTimer(1 * time.Second)
	select {
	case <-t.C:
		catAddrs, _, err := cw.cc.Catalog().Service(cw.cr.ServiceName, "", nil)
		if err == nil {
			addrNow := make([]string, 0)
			for _, cs := range catAddrs {
				// I'm not sure if ServiceAddress always has a value
				// If it is "", use Address
				sd := cs.ServiceAddress
				if sd == "" {
					sd = cs.Address
				}
				// addr should like: 127.0.0.1:8001
				addrNow = append(addrNow, fmt.Sprintf("%s:%d", sd, cs.ServicePort))
			}

			deleted := diff(cw.addrs, addrNow)
			for _, addr := range deleted {
				update := &naming.Update{Op: naming.Delete, Addr: addr}
				updates = append(updates, update)
			}

			added := diff(addrNow, cw.addrs)
			for _, addr := range added {
				update := &naming.Update{Op: naming.Add, Addr: addr}
				updates = append(updates, update)
			}

			// update addrs
			cw.addrs = addrNow

			// if updates, return
			if len(updates) != 0 {
				return updates, nil
			}
		}

		t.Reset(1 * time.Second)
	}

	return nil, nil
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
