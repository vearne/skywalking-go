package core

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sync"
)

type PromWrapper struct {
	registry   *prometheus.Registry
	storehouse *sync.Map
}

func NewPromWrapper() *PromWrapper {
	return &PromWrapper{registry: nil, storehouse: &sync.Map{}}
}

func (p *PromWrapper) NewCounterVec(name, help string, labelNames []string) interface{} {
	val, ok := p.storehouse.Load(name)
	if ok {
		return val
	}

	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labelNames)

	if p.registry != nil {
		err := p.registry.Register(cv)
		if err != nil {
			log.Printf("[----1----]registry.Register, error:%v\n", err)
		}
		w := NewCounterVecWrapper(cv)
		p.storehouse.Store(name, w)
		return w
	} else {
		log.Println("[----2----]PromWrapper-NewCounterVec, registry is nil")
		return nil
	}
}

func (p *PromWrapper) SetRegistry(registry interface{}) {
	if p.registry == nil {
		p.registry = registry.(*prometheus.Registry)
	}
}

type CounterVecWrapper struct {
	cv *prometheus.CounterVec
}

func NewCounterVecWrapper(cv *prometheus.CounterVec) *CounterVecWrapper {
	return &CounterVecWrapper{cv: cv}
}

func (c *CounterVecWrapper) With(labels map[string]string) interface{} {
	return c.cv.With(labels)
}
