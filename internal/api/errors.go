package api

import (
	"fmt"

	"github.com/tktip/flyvo-api/pkg/rpc"
)

func errorWrongResponseCodeFlyvoRPC(response rpc.Generic) error {
	return fmt.Errorf("wrong response from flyvo (%d): %s", response.Status, response.Body)
}

func errorWrongResponseCodeVigiloHTTP(body []byte, status string) error {
	return fmt.Errorf("wrong response from vigilo (%s): %s", status, body)
}
