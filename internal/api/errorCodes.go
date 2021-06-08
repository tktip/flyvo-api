package api

import "github.com/gin-gonic/gin"

type errorCode int

const (
	//CodeNotFound - resource not found
	CodeNotFound errorCode = iota

	//CodeAlreadyRegistered - participant already registered
	CodeAlreadyRegistered

	//CodeRedisError - error communicating with redis
	CodeRedisError

	//CodeQRError - error creating qr code
	CodeQRError

	//CodeNotAuthenticated - not authenticated
	CodeNotAuthenticated

	//CodeForbidden - user does not have access to resource
	CodeForbidden

	//CodeBadRequest - bad data provided by client
	CodeBadRequest

	//CodeConnectionError - failed to connect to something
	CodeConnectionError

	//CodeUnexpectedResponse - unexpected response (e.g. from AD)
	CodeUnexpectedResponse

	//CodeInternalErrorGeneral - for general errors.
	CodeInternalErrorGeneral
)

func codedErrorResponse(msg string, code errorCode) gin.H {
	return gin.H{
		"error": msg,
		"code":  code,
	}
}
