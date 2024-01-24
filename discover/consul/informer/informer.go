// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
