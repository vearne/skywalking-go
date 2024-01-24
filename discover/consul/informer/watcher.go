package informer

import (
	"github.com/fatih/structs"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	slog "github.com/vearne/simplelog"
	"math/rand"
	"time"
)

type ServiceWatcher struct {
	ServiceName  string
	DC           string
	Token        string
	Plan         *watch.Plan // individual monitoring for each service
	ServiceState *State      // latest state
	ConsulAddr   string
	Tags         []string
}

func NewServiceWatcher(service string, c *ConsulInfo, tags []string, ch chan<- *StateChange) (*ServiceWatcher, error) {
	slog.Info("NewServiceWatcher, dc:%v, service:%v", c.DC, service)

	var w ServiceWatcher
	var err error
	w.ServiceName = service
	w.DC = c.DC
	w.Tags = tags
	// initial state
	w.ServiceState = &State{ServiceEntrys: make([]*consulapi.ServiceEntry, 0), T: time.Now(), Index: 0}

	N := len(c.Addresses)
	w.ConsulAddr = c.Addresses[rand.Intn(N)]

	param := PlanParam{
		Type:        "service",
		Service:     service,
		PassingOnly: true,
		DC:          w.DC,
		Token:       c.Token,
		Stale:       false,
		Tag:         tags,
	}

	w.Plan, err = watch.Parse(structs.Map(&param))
	if err != nil {
		return nil, err
	}

	w.Plan.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		case []*consulapi.ServiceEntry:
			// state changed
			if idx != w.ServiceState.Index {
				newState := State{ServiceEntrys: d, T: time.Now(), Index: idx}
				ch <- &StateChange{
					NewState:  newState,
					DC:        w.DC,
					Service:   w.ServiceName,
					LastIndex: w.ServiceState.Index,
				}
				// modify the current state to the latest state
				w.ServiceState = &newState
			}
		default:
			slog.Error("unknown data type, type:%v, data:%v", d, data)
		}
	}

	return &w, nil
}

func (w *ServiceWatcher) Run() error {
	slog.Info("ServiceWatcher-Run(), dc:%v, service:%v", w.DC, w.ServiceName)
	err := w.Plan.Run(w.ConsulAddr)
	if err != nil {
		slog.Error("ServiceWatcher, dc:%v, service:%v, error:%v", w.DC, w.ServiceName, err)
		return err
	}
	return nil
}

func (w *ServiceWatcher) Stop() {
	slog.Info("ServiceWatcher-Stop(), dc:%v, service:%v", w.DC, w.ServiceName)
	w.Plan.Stop()
}

type PlanParam struct {
	Type        string   `structs:"type"`
	Service     string   `structs:"service"`
	PassingOnly bool     `structs:"passingonly"`
	DC          string   `structs:"datacenter,omitempty"`
	Token       string   `structs:"token,omitempty"`
	Tag         []string `structs:"tag,omitempty"`
	Stale       bool     `structs:"stale,omitempty"`
}
