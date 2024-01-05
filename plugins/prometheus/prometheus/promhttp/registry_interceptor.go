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

package prometheus

import (
	"github.com/apache/skywalking-go/plugins/core/log"
	"github.com/apache/skywalking-go/plugins/core/operator"
	"github.com/apache/skywalking-go/plugins/core/prom"
	"github.com/prometheus/client_golang/prometheus"
)

type RegistryInterceptor struct {
}

func (h *RegistryInterceptor) BeforeInvoke(invocation operator.Invocation) error {

	return nil
}

func (h *RegistryInterceptor) AfterInvoke(invocation operator.Invocation, result ...interface{}) error {
	log.Infof("-----register prometheus------, %p\n", prometheus.DefaultRegisterer)
	prom.SetRegistry(prometheus.DefaultRegisterer)
	return nil
}
