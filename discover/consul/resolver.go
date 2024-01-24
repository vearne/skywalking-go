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

package consul

import (
	"errors"
	"fmt"
	"github.com/apache/skywalking-go/discover/consul/informer"
	"github.com/hashicorp/consul/api"
	slog "github.com/vearne/simplelog"
	"google.golang.org/grpc/resolver"
	"os"
	"strings"
	"sync"
)

const (
	resolverSchemeConsul = "consul"
)

func init() {
	slog.Info("resolver.Register consulBuilder")
	resolver.Register(newConsulBuilder())
}

// consulBuilder builds the address resolver for gRPC dialer.
type consulBuilder struct {
}

// newConsulBuilder is a constructor.
func newConsulBuilder() *consulBuilder {
	return &consulBuilder{}
}

// Builds the consul address resolver for gRPC.
// SW_AGENT_REPORTER_GRPC_BACKEND_SERVICE=consul://127.0.0.1:8500/skywalking-oap-server
// consul://127.0.0.1:8500/service?tag=product&tag=project%3Dfake
func (c consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	slog.Info("rawURL: %v", target.URL.String())

	consulAddr, serviceName, tags, err := parseTarget(target)
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	slog.Info("parseTarget, consulAddr:%v, serviceName:%v, tags:%v",
		consulAddr, serviceName, tags)

	// 1) get consul address
	consulList := make([]informer.ConsulInfo, 0)
	// 2) get dc information
	dcList := GetDCList()
	if len(dcList) <= 0 {
		return nil, errors.New("The DC list must be specified using the environment variable CONSUL_DISCOVER_DC_LIST")
	}
	for _, dc := range dcList {
		info := informer.ConsulInfo{Addresses: []string{consulAddr}, DC: dc}
		consulList = append(consulList, info)
	}

	in := informer.NewInformer(consulList, []string{serviceName}, tags)
	cr := newConsulResolver(cc, in)
	err = cr.watch()
	if err != nil {
		return nil, fmt.Errorf("watch: %w", err)
	}

	return cr, nil
}

// CONSUL_DISCOVER_DC_LIST=beijing,shanghai
func GetDCList() []string {
	str := os.Getenv("CONSUL_DISCOVER_DC_LIST")
	return strings.Split(str, ",")
}

// parses the target and returns service name and tag.
// consul://127.0.0.1:8500/service?tag=product&tag=project%3Dfake
func parseTarget(target resolver.Target) (consulAddr, serviceName string, tags []string, err error) {
	u := target.URL

	consulAddr = u.Host
	serviceName = strings.Trim(u.Path, "/")
	tags = u.Query()["tag"]
	err = nil
	return
}

// Scheme returns the consul resolver scheme.
func (c consulBuilder) Scheme() string {
	return resolverSchemeConsul
}

type consulResolver struct {
	cc              resolver.ClientConn
	in              *informer.Informer
	stateChangeChan chan *informer.StateChange
	state           *GlobalState
}

func newConsulResolver(cc resolver.ClientConn, in *informer.Informer) *consulResolver {
	return &consulResolver{
		cc:    cc,
		in:    in,
		state: NewGlobalState(),
	}
}

func (c *consulResolver) onServiceChanged(entries []*api.ServiceEntry) {
	addresses := make([]resolver.Address, len(entries))
	for i, e := range entries {
		addresses[i] = resolver.Address{
			Addr: fmt.Sprintf("%v:%v", e.Service.Address, e.Service.Port),
		}
	}

	err := c.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
	if err != nil {
		slog.Error("update gRPC consul resolver addresses, error: %v", err)
	}
}

func (c *consulResolver) watch() error {
	c.stateChangeChan = c.in.Watch()
	go func() {
		for sc := range c.stateChangeChan {
			slog.Debug("stateChange, dc:%v, service:%v, LastIndex:%v, Index:%v, len(ServiceEntrys):%v",
				sc.DC, sc.Service, sc.LastIndex, sc.NewState.Index, len(sc.NewState.ServiceEntrys))

			c.state.Set(sc.DC, sc.NewState.ServiceEntrys)
			// update connection pool
			c.onServiceChanged(c.state.GetGlobalState())
		}
	}()

	return nil
}

// ResolveNow we don't need to anything here because all addresses are updated on change in consul.
func (c *consulResolver) ResolveNow(options resolver.ResolveNowOptions) {}

// Close is an interface method. Gracefully closes the consulResolver.
func (c *consulResolver) Close() {
	c.in.Stop()
	close(c.stateChangeChan)
}

// save a list of instances in multiple data centers
type GlobalState struct {
	// dc -> []*api.ServiceEntry
	inner map[string][]*api.ServiceEntry
	m     sync.RWMutex
}

func NewGlobalState() *GlobalState {
	var g GlobalState
	g.inner = make(map[string][]*api.ServiceEntry)
	return &g
}

func (g *GlobalState) Set(dc string, entrys []*api.ServiceEntry) {
	g.m.Lock()
	defer g.m.Unlock()
	g.inner[dc] = entrys
}

func (g *GlobalState) GetGlobalState() []*api.ServiceEntry {
	g.m.RLock()
	defer g.m.RUnlock()
	// remove duplicates
	exist := make(map[string]struct{})
	result := make([]*api.ServiceEntry, 0)
	for dc, entrys := range g.inner {
		slog.Debug("GetGlobalState, dc:%v, count:%v", dc, len(entrys))
		for _, entry := range entrys {
			key := fmt.Sprintf("%v:%v", entry.Service.Address, entry.Service.Port)
			if _, ok := exist[key]; !ok {
				result = append(result, entry)
				exist[key] = struct{}{}
			}
		}
	}

	return result
}
