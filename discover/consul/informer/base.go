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
	consulapi "github.com/hashicorp/consul/api"
	"time"
)

type StateChange struct {
	NewState  State
	DC        string
	Service   string
	LastIndex uint64
}

type State struct {
	ServiceEntrys []*consulapi.ServiceEntry
	T             time.Time
	Index         uint64
}

type ConsulInfo struct {
	Addresses []string `json:"addresses"`
	DC        string   `json:"dc"`
	Token     string   `json:"token"`
}
