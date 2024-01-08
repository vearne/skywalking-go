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

package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/apache/skywalking-go/plugins/core/log"
	"github.com/apache/skywalking-go/plugins/core/operator"
	"github.com/apache/skywalking-go/plugins/core/prom"
	"github.com/apache/skywalking-go/plugins/core/tracing"
)

type HTTPInterceptor struct {
}

func (h *HTTPInterceptor) BeforeInvoke(invocation operator.Invocation) error {
	context := invocation.Args()[0].(*gin.Context)
	s, err := tracing.CreateEntrySpan(
		fmt.Sprintf("%s:%s", context.Request.Method, context.Request.URL.Path), func(headerKey string) (string, error) {
			return context.Request.Header.Get(headerKey), nil
		},
		tracing.WithLayer(tracing.SpanLayerHTTP),
		tracing.WithTag(tracing.TagHTTPMethod, context.Request.Method),
		tracing.WithTag(tracing.TagURL, context.Request.Host+context.Request.URL.Path),
		tracing.WithComponent(5006))
	if err != nil {
		return err
	}
	invocation.SetContext(s)
	return nil
}

func (h *HTTPInterceptor) AfterInvoke(invocation operator.Invocation, result ...interface{}) error {
	if invocation.GetContext() == nil {
		return nil
	}
	context := invocation.Args()[0].(*gin.Context)
	span := invocation.GetContext().(tracing.Span)
	span.Tag(tracing.TagStatusCode, fmt.Sprintf("%d", context.Writer.Status()))
	if len(context.Errors) > 0 {
		span.Error(context.Errors.String())
	}
	span.End()

	// add log
	log.Infof("url:%v", context.Request.URL.Path)

	// add metrics
	httpReqTotal := prom.GetOrNewCounterVec(
		"gin_requests_total",
		"Total number of gin HTTP requests made",
		[]string{"method", "path", "status"},
	)

	httpReqTotal.With(map[string]string{
		"method": context.Request.Method,
		"path":   context.Request.URL.Path,
		"status": fmt.Sprintf("%d", context.Writer.Status()),
	}).(prom.Counter).Inc()

	return nil
}
