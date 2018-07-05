package brpc

import (
	"context"
)

type CallId struct {
}

type Controller struct {
	ctx    context.Context
	callId CallId
}
