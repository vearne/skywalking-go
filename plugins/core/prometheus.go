package core

import (
	"github.com/apache/skywalking-go/plugins/core/operator"
	"github.com/prometheus/client_golang/prometheus"
)

type PromWrapper struct {
	registry *prometheus.Registry
}

func NewPromWrapper() *PromWrapper {
	return &PromWrapper{registry: nil}
}

func (p *PromWrapper) NewCounterVec(name, help string, labelNames []string) operator.CounterVec {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labelNames)

	if p.registry != nil {
		p.registry.Register(cv)
		return NewCounterVecWrapper(cv)
	} else {
		op := operator.GetOperator()
		if op == nil {
			return nil
		}
		op.Logger().(operator.LogOperator).Error("PromWrapper-NewCounterVec, registry is nil")
		return nil
	}
}

func (p *PromWrapper) SetRegistry(registry interface{}) {
	p.registry = registry.(*prometheus.Registry)
}

type CounterVecWrapper struct {
	cv *prometheus.CounterVec
}

func NewCounterVecWrapper(cv *prometheus.CounterVec) *CounterVecWrapper {
	return &CounterVecWrapper{cv: cv}
}

func (c *CounterVecWrapper) With(labels map[string]string) operator.Counter {
	return c.cv.With(labels)
}
