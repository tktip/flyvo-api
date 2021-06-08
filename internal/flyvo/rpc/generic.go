package rpc

import (
	"context"
	"net/http"

	"github.com/tktip/flyvo-api/pkg/rpc"
)

//revive:disable:unused-receiver

const (
	ping = "ping"
)

func (s *Server) ping(_ context.Context, _ rpc.Generic) *rpc.Generic {
	return &rpc.Generic{
		Body:   []byte(`Pong`),
		Status: http.StatusOK,
	}
}
