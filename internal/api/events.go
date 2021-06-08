package api

//revive:disable:line-length-limit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/pkg/flyvo"
	"github.com/tktip/flyvo-api/pkg/rpc"

	"github.com/gin-gonic/gin"

	"github.com/tktip/flyvo-api/internal/errorhandler"
)

const (
	day          = 24 * time.Hour
	maxDateRange = 14 * day
	layout       = "02.01.2006"
)

// getEventsForTeacher returns events where the user is an expected participant
// @Summary Returns events where the user is an expected participant
// @Description Returns events where the user is an expected participant
// @Produce application/json
// @Success 200 {string} string "json user object"
// revive:disable-line:line-length
// @Failure 400 {string} string "If missing or bad params, or if auth proxy user-data-b64 header is missing (e.g. auth proxy circumvented)"
// @Failure 500 {string} string "On any other error (e.g. rpc)"
// @Router /event/retrieve [GET]
func (s *Server) getEventsForTeacher(c *gin.Context) {
	if !isTeacher(c) {
		return
	}

	DateFromInclusive := c.Param("from")
	DateToInclusive := c.Param("to")

	start, err := time.Parse(layout, DateFromInclusive)
	if err != nil {
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			"Bad start time value",
			CodeBadRequest,
		))
		return
	}

	end, err := time.Parse(layout, DateToInclusive)
	if err != nil {
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			"Bad end time value",
			CodeBadRequest,
		))
		return
	}

	if end.Before(start) {
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			"End before start",
			CodeBadRequest,
		))
		return
	}

	if end.Sub(start) > maxDateRange {
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			fmt.Sprintf("Range exceeds maximum (%s)", maxDateRange),
			CodeBadRequest,
		))
		return
	}

	req := flyvo.GetCoursesRequest{
		//TeacherID: extractVismaID(person.Email),
		FromDate: start,
		ToDate:   end,
	}

	gen := &rpc.Generic{
		Path: rpc.PathGetTeacherCourses,
	}

	gen.Body, err = json.Marshal(req)
	if err != nil {
		logrus.Errorf("Unable to marshal request body: %s", err.Error())
		errorhandler.HandleRPCError(c, err)
		return
	}

	response, err := s.RPC.WaitForClientsideProcessing(gen, time.Second*15)

	if err != nil {
		logrus.Errorf("Failed during clientside processing: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"Error communicating with FlyVO",
			CodeConnectionError,
		))
		errorhandler.HandleRPCError(c, err)
		return
	}

	if response.Status != 200 {
		logrus.Warnf("Unexpected response code from FlyVO: %d", response.Status)
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"Unexpected response from FlyVO",
			CodeUnexpectedResponse,
		))
		return
	}

	c.Header("content-type", "application/json")
	c.String(int(response.Status), string(response.Body))
}

// registerParticipation register student to an activity
// @Summary Register participation
// @Description Registers a person as having participated in an event
// @Produce application/json
// @Param activityId query string true "activity a person wants to participate in"
// @Success 200 {string} string "If participant was successfully registered in event"
// @Failure 422 {string} string "No activity mapped to provided id"
// @Failure 500 {string} string "On any unexpected error (e.g. unable to connect to Redis)"
// @Router /event/participate/ [GET]
func (s *Server) registerParticipation(c *gin.Context) {

	person, ok := getPersonObject(c)
	if !ok {
		return
	}

	participationID := c.Query("participationId")
	activityID, err := s.Redis.GetStringValue(participationID)
	if err != nil {
		logrus.Errorf("Failed to read activity ID from redis: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"unable to retrieve activity data from db",
			CodeRedisError,
		))
		return
	} else if activityID == "" {
		c.JSON(http.StatusUnprocessableEntity, codedErrorResponse(
			"Activity id not found",
			CodeNotFound,
		))
		return
	}

	//Check if user already participation
	pString := fmt.Sprintf("participation-%s-%s", activityID, person.Email)
	val, err := s.Redis.GetValue(pString)
	if err != nil {
		logrus.Errorf("Participation register get error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"failed to check participation status in db",
			CodeRedisError,
		))
		return
	} else if val != nil {
		logrus.Debugf("Event was already registered: %s", pString)
		c.JSON(http.StatusBadRequest, codedErrorResponse(
			"already registered",
			CodeAlreadyRegistered,
		))
		return
	}

	//If not, register participation
	err = s.Redis.WriteValue(pString, time.Now().String())
	if err != nil {
		logrus.Errorf("Participation register write error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"failed to register participation in db",
			CodeRedisError,
		))
		return
	}

	c.Status(http.StatusNoContent)
}
