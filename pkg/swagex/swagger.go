package swagex

//revive:disable

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

// global swagger details
// @title Flyvo API Swagger
// @version 1.0
// @description Swagger for FLYVO API
// @description FLYVO API provides access to FLYVO data from Visma.
// @description the Api is used to provide access to secure zone data at Visma,
// @description such as event details, participation and so on.
// @description Communication with Visma is performed vis RPC calls where the API is an
// @description RPC server. As such, the calls to Visma from frontend are faux synchronous.
// @description The connection is kept open while the server waits for data from client at Visma side.

// @contact.name TIP

// @host localhost:7070
// @BasePath /

// @securityDefinitions.basic BasicAuth

var (
	swaggerDoc = []byte("{}")
)

func init() {
	swaggerLoc := os.Getenv("SWAGGER_LOCATION")
	if swaggerLoc == "" {
		swaggerLoc = "/swagger.json"
	}

	var err error
	swaggerDoc, err = ioutil.ReadFile(swaggerLoc)
	if err != nil || swaggerDoc == nil {
		logrus.Warn("Could not read swagger doc: " + err.Error())
		swaggerDoc = []byte("{}")
	}
}

// SwaggerEndpoint - returns swagger  doc
// @Summary Swagger doc.
// @Description Returns the swagger doc.
// @Produce application/json
// @Success 200 {string} string "The swagger json."
// @Router /api-doc [get]
func SwaggerEndpoint(c *gin.Context) {
	c.Writer.Header().Set("content-type", "application/json; charset=UTF-8")
	c.Writer.Write(swaggerDoc)
}
