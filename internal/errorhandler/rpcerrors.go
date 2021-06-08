package errorhandler

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

//revive:disable:unnecessary-stmt

//HandleRPCError responds with a proper response based on error from rpc.
func HandleRPCError(c *gin.Context, err error) {
	switch err.Error() {
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"errorCode": 3000,
		})
		logrus.Errorf("Rpc error: %s", err.Error())
	}
}
