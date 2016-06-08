package consul 

import (
	"errors"
	"fmt"
	"log"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/wothing/wonaming/lib"
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
					updates := lib.GenUpdates([]string{}, addrs)
					return updates, nil
				}
				log.Println("no such service found in consul, retry")
			}
		}
	}

	addrs, li, err := cw.queryConsul(&consul.QueryOptions{WaitIndex: cw.li})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("consul query error: %v", err))
	}

	updates := lib.GenUpdates(cw.addrs, addrs)

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

