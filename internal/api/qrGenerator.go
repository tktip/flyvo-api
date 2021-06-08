package api

//revive:disable:line-length-limit

import (
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// generateParticipationID generate participation ID for event
// @Summary Generate participation ID for event
// @Description Generates participation ID for event based on activity ID
// @Produce application/json
// @Param activityId query string true "id of an existing event"
// @Success 200 {string} string "A participation ID in a json structure: {'participationId':ID'} "
// @Failure 403 {string} string "If unauthorized (not teacher in GCE)"
// @Failure 500 {string} string "On any other error"
// @Router /generate/participationId [GET]
func (s *Server) generateParticipationID(c *gin.Context) {

	if !isTeacher(c) {
		return
	}

	activityID := c.Query("activityId")

	participationPIN := ""
	counter := 0
	numberOfRunes := 4
	for {
		participationPIN = randStringRunes(numberOfRunes)
		response, err := s.Redis.GetValue(participationPIN)
		if err != nil {
			logrus.Errorf("Failed to get participation Id value from redis: %s", err.Error())
			c.JSON(http.StatusInternalServerError, codedErrorResponse(
				"failed to create PIN",
				CodeRedisError,
			))
			return
		}

		if response == nil {
			break
		}

		if counter == 10000 {
			numberOfRunes++
			counter = 0
		}
		counter = counter + 1
	}

	err := s.Redis.WriteValue(participationPIN, activityID)
	if err != nil {
		logrus.Errorf("Failed to write participation Id activity ID pair to redis: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"failed to create PIN",
			CodeRedisError,
		))
		return
	}

	err = s.Redis.WriteValue("activity-"+activityID, activityID)
	if err != nil {
		logrus.Errorf("Failed to register activity '%s' as having occurred: %s ", activityID, err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"failed to register activity",
			CodeRedisError,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participationId": participationPIN,
	})
}

// generateQrCode generate QR code based on participationId
// @Summary Create QR code
// @Description Generates a QR code image encoded in base64
// @Produce application/json
// @Param participationId query string true "id displayed to participants in physical form and QR code"
// @Success 200 {string} string "A base 64 encoded image of QR code"
// @Failure 403 {string} string "If unauthorized (not teacher in GCE)"
// @Failure 422 {string} string "If participation ID not found"
// @Failure 500 {string} string "On any unexpected error (e.g. unable to connect to Redis)"
// @Router /generateQrCode [GET]
func (s *Server) generateQrCode(c *gin.Context) {
	participationID := c.Query("participationId")

	if !isTeacher(c) {
		return
	}

	client := &http.Client{}

	activityID, err := s.Redis.GetValue(participationID)
	if err != nil {
		logrus.Errorf("Failed to retrieve activity ID from redis: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"Failed to get activity from db",
			CodeRedisError,
		))
		return
	} else if activityID == nil {
		logrus.Debugf("No activity mapped to id '%s'", participationID)
		c.JSON(http.StatusUnprocessableEntity, codedErrorResponse(
			"No activity mapped to that ID",
			CodeNotFound,
		))
		return
	}

	req, err := http.NewRequest("GET", s.QrURL, nil)
	if err != nil {
		logrus.Errorf("Failed to create qr code request: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"Unable to generate QR code",
			CodeQRError,
		))
		return
	}

	q := req.URL.Query()
	q.Add("data", s.ParticipantURL+"?id="+participationID)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Failed to retrieve qr code: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"Failed to retrieve QR code",
			CodeQRError,
		))
		return
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Failed to read qr code response: %s", err.Error())
		c.JSON(http.StatusInternalServerError, codedErrorResponse(
			"Failed to retrieve QR code",
			CodeQRError,
		))
		return
	}

	// logrus.Info(string(bytes))
	c.Data(http.StatusOK, resp.Header.Get("content-type"), bytes)
}
