package api

//revive:disable:line-length-limit
//revive:disable:cyclomatic

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/pkg/flyvo"
	"github.com/tktip/flyvo-api/pkg/rpc"

	"github.com/gin-gonic/gin"

	"github.com/tktip/flyvo-api/internal/errorhandler"
	"github.com/tktip/flyvo-api/internal/structs"
)

var (
	getEvent = "/events/"
)

const (
	layoutISO = "2006-01-02"
)

func (s *Server) getSickleaves(c *gin.Context) {
	user, ok := getPersonObject(c)
	if !ok {
		return
	}

	to, err := time.Parse(layout, c.Param("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, codedErrorResponse("bad to time value", CodeBadRequest))
		return
	}
	layoutISO := "02012006"
	formatDate := to.Format(layoutISO)

	req := flyvo.GetSickLeavesRequest{
		VismaID: extractVismaID(user.Email),
		ToDate:  formatDate,
	}

	gen := rpc.Generic{
		Path: rpc.PathGetSickLeaves,
	}

	gen.Body, err = json.Marshal(req)
	if err != nil {
		logrus.Errorf("Unable to marshal request body: %s", err.Error())
		errorhandler.HandleRPCError(c, err)
		return
	}

	response, err := s.RPC.WaitForClientsideProcessing(
		&gen,
		time.Second*15,
	)

	if err != nil {
		logrus.Errorf("Failed during clientside processing: %s", err.Error())
		errorhandler.HandleRPCError(c, err)
		return
	}

	if response.Status != 200 {
		logrus.Errorf("Unexpected response code from flyvo: %d", response.Status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     errorWrongResponseCodeFlyvoRPC(response).Error(),
			"errorCode": 1,
		})
		return
	}

	c.Header("content-type", "application/json")
	c.String(int(response.Status), string(response.Body))
}

// getAbsenceCount returns registered absence count for user
// @Summary Returns amount of previous absences registered
// @Description Returns amount of previous absences registered for a specific user
// @Produce text/plain
// @Success 200 {string} string "the absence count for the user"
// @Failure 400 {string} string "If auth proxy user-data-b64 header is missing (e.g. auth proxy circumvented)"
// @Failure 500 {string} string "On any other error (e.g. rpc)"
// @Router /absence/count [GET]
func (s *Server) getAbsenceCount(c *gin.Context) {
	user, ok := getPersonObject(c)
	if !ok {
		return
	}

	from, err := time.Parse(layout, c.Param("from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, codedErrorResponse("bad from time value", CodeBadRequest))
		return
	}

	to, err := time.Parse(layout, c.Param("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, codedErrorResponse("bad to time value", CodeBadRequest))
		return
	}

	req := flyvo.GetUnauthorizedAbsenceRequest{
		VismaID:  extractVismaID(user.Email),
		FromDate: from,
		ToDate:   to,
	}

	gen := rpc.Generic{
		Path: rpc.PathGetAbsences,
	}

	gen.Body, err = json.Marshal(req)
	if err != nil {
		logrus.Errorf("Unable to marshal request body: %s", err.Error())
		errorhandler.HandleRPCError(c, err)
		return
	}

	response, err := s.RPC.WaitForClientsideProcessing(
		&gen,
		time.Second*15,
	)

	if err != nil {
		logrus.Errorf("Failed during clientside processing: %s", err.Error())
		errorhandler.HandleRPCError(c, err)
		return
	}

	if response.Status != 200 {
		logrus.Errorf("Unexpected response code from flyvo: %d", response.Status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     errorWrongResponseCodeFlyvoRPC(response).Error(),
			"errorCode": 1,
		})
		return
	}

	c.Header("content-type", "application/json")
	c.String(int(response.Status), string(response.Body))
}

// registerSickleave registers absence to an event
// @Summary Registers a user as absent in provided event
// @Description Registers a user as absent in provided event
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string "OK, user was registered as absent. Current absence count returned."
// @Failure 400 {string} string "If auth proxy user-data-b64 header is missing (e.g. auth proxy circumvented)"
// @Failure 500 {string} string "On any other error (e.g. rpc)"
// @Router /absence/register [POST]
func (s *Server) registerSickleave(c *gin.Context) {
	person, ok := getPersonObject(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	absence := structs.EventSickLeave{}
	err := c.BindJSON(&absence)
	if err != nil {
		logrus.Errorf("Failed to bind eventsickleave: %s", err.Error())
		c.JSON(http.StatusUnprocessableEntity, codedErrorResponse("bad request body",
			CodeBadRequest,
		))
		return
	}

	if absence.AbsenceCode == "" ||
		absence.Start == nil ||
		absence.End == nil {
		logrus.Errorf("Failed to bind eventsickleave: " +
			"Missing absence code, start or end")
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			"missing absence type, or absence start/end",
			CodeBadRequest,
		))
		return
	}

	if absence.AbsenceCode == "0" {
		absence.AbsenceCode = "E"
	} else if absence.AbsenceCode == "1" {
		absence.AbsenceCode = "A"
	} else {
		logrus.Errorf("Invalid absence code '%s'", absence.AbsenceCode)
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			"invalid absence code",
			CodeBadRequest,
		))
		return
	}

	logrus.Info(absence.Start.Format(layoutISO))
	logrus.Info(absence.End.Format(layoutISO))

	req := flyvo.RegisterSickLeave{
		VismaID:  extractVismaID(person.Email),
		Code:     absence.AbsenceCode,
		FromDate: absence.Start.Format(layoutISO),
		ToDate:   absence.End.Format(layoutISO),
	}

	generic := &rpc.Generic{
		Path: rpc.PathRegisterSickLeave,
	}

	generic.Body, err = json.Marshal(&req)
	if err != nil {
		errorhandler.HandleRPCError(c, err)
		logrus.Errorf("Failed to marshal generic tiprpc request: %s", err.Error())
		return
	}

	response, err := s.RPC.WaitForClientsideProcessing(
		generic,
		time.Second*15,
	)

	if err != nil {
		logrus.Errorf("Failed during clientside processing: %s", err.Error())
		errorhandler.HandleRPCError(c, err)
		return
	}

	if response.Status != http.StatusOK && response.Status != http.StatusNoContent {
		logrus.Errorf("Unexpected response code from flyvo: %d", response.Status)
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			errorWrongResponseCodeFlyvoRPC(response).Error(),
			CodeUnexpectedResponse,
		))
		return
	}

	c.Status(int(response.Status))
}
