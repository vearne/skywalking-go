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

package prom

import (
	"github.com/apache/skywalking-go/plugins/core/operator"
	"github.com/apache/skywalking-go/plugins/core/tools"
)

type Counter interface {
	Inc()
	Add(float64)
}

var storehouse tools.SyncMap

func init() {
	storehouse = tools.NewSyncMap()
}

func SetRegistry(registry interface{}) {
	op := operator.GetOperator()
	if op == nil {
		return
	}
	op.PromMetrics().(operator.PromOperator).SetRegistry(registry)
	return
}

func GetOrNewCounterVec(name, help string, labelNames []string) operator.CounterVec {
	op := operator.GetOperator()
	if op == nil {
		return nil
	}

	if val, ok := storehouse.Get(name); ok {
		return val.(operator.CounterVec)
	}

	cv := op.PromMetrics().(operator.PromOperator).NewCounterVec(name, help, labelNames)
	storehouse.Put(name, cv)
	return cv.(operator.CounterVec)
}
