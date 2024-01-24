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
