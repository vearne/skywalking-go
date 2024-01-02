package prometheus

import (
	"github.com/apache/skywalking-go/plugins/core/log"
	"github.com/apache/skywalking-go/plugins/core/operator"
)

type RegistryInterceptor struct {
}

func (h *RegistryInterceptor) BeforeInvoke(invocation operator.Invocation) error {

	return nil
}

func (h *RegistryInterceptor) AfterInvoke(invocation operator.Invocation, result ...interface{}) error {
	log.Infof("-----register prometheus------")
	return nil
}
