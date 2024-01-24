package informer

import (
	slog "github.com/vearne/simplelog"
	"golang.org/x/sync/errgroup"
)

type Informer struct {
	consuls      []ConsulInfo
	serviceNames []string

	group    errgroup.Group
	watchers []*ServiceWatcher
	tags     []string
}

func NewInformer(consuls []ConsulInfo, serviceNames []string, tags []string) *Informer {
	var informer Informer
	informer.consuls = consuls
	informer.serviceNames = serviceNames
	informer.tags = tags
	informer.watchers = make([]*ServiceWatcher, 0)
	return &informer
}

func (informer *Informer) Watch() chan *StateChange {
	ch := make(chan *StateChange, 100)
	for idx := range informer.consuls {
		c := informer.consuls[idx]
		for _, service := range informer.serviceNames {
			watcher, err := NewServiceWatcher(service, &c, informer.tags, ch)
			if err != nil {
				slog.Error("create ServiceWatcher, dc:%v, service:%v, error:%v", c.DC, service, err)
				continue
			}

			informer.watchers = append(informer.watchers, watcher)
			informer.group.Go(func() error {
				return watcher.Run()
			})
		}
	}

	return ch
}

func (informer *Informer) Stop() {
	for _, watcher := range informer.watchers {
		watcher.Stop()
	}
	err := informer.group.Wait()
	slog.Error("Informer.Stop(), error:%v", err)
}
