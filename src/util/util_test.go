package util

import (
	"fmt"
	"testing"
)

func TestExtractServiceAndMethod(t *testing.T) {
	var service = "ServiceHello"
	var method = "MethodEcho"
	var serviceMethod = fmt.Sprintf("%s.%s", service, method)

	var s, m, err = ExtractServiceAndMethod(serviceMethod)
	if err != nil ||
		s != service ||
		m != method {
		t.Errorf("ExtractServiceAndMethod(%s) = %s, %s, %v, expect %s, %s",
			serviceMethod, s, m, err, service, method)
	}
}
