package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/pkg/flyvo"
	"github.com/tktip/flyvo-api/pkg/rpc"
)

type eventSet map[string]map[string]bool
type gcalResp struct {
	Events struct {
		Items []struct {
			ID        string `json:"id"`
			Attendees []struct {
				Email string `json:"email"`
			} `json:"attendees"`
			Status string `json:"status"`
			Start  struct {
				DateTime time.Time `json:"dateTime"`
			} `json:"start"`
			End struct {
				DateTime time.Time `json:"dateTime"`
			} `json:"end"`
		} `json:"items"`
	} `json:"events"`
}

type gEvent struct {
	ID        string `json:"id"`
	Attendees []struct {
		Email string `json:"email"`
	} `json:"attendees"`
	Status string `json:"status"`
	Start  struct {
		DateTime time.Time `json:"dateTime"`
	} `json:"start"`
	End struct {
		DateTime time.Time `json:"dateTime"`
	} `json:"end"`
}

type gcalSingleResp struct {
	Event gEvent `json:"event"`
}

//TimeWithinLimits - checks that the start and end times are within the bounds of limit
func TimeWithinLimits(start, end, startLimit, endLimit time.Time) bool {
	startAfterEnd := endLimit.After(start)
	startBeforeStart := start.Before(startLimit)
	endAfterEnd := endLimit.After(end)
	endBeforeStart := end.Before(startLimit)
	return !(startAfterEnd || startBeforeStart || endAfterEnd || endBeforeStart)
}

var (
	invalidChars = regexp.MustCompile(`([a-z0-9])*`)
)

func sanitizeCalendarID(ID string) string {
	return strings.Join(invalidChars.FindAllString(strings.ToLower(ID), -1), "")
}

func (s *Server) getTodaysEvents() (eventSet, error) {
	list, err := s.Redis.GetList("activity-*")
	if err != nil {
		return nil, err
	}

	var googleEvents []gEvent
	for _, key := range list {
		activityID := strings.Split(key, "-")[2]
		url := s.GcalURL + eventGet + sanitizeCalendarID(activityID)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logrus.Warnf("Unexpected status for activity '%s' from calendar: %s",
				activityID,
				resp.Status,
			)
			continue
		}
		result := gcalSingleResp{}
		b, _ := ioutil.ReadAll(resp.Body)
		err = json.NewDecoder(bytes.NewReader(b)).Decode(&result)
		if err != nil {
			return nil, err
		}
		googleEvents = append(googleEvents, result.Event)
	}

	eSet := make(eventSet)
	for _, event := range googleEvents {
		if event.Status != "confirmed" {
			continue
		}

		eSet[event.ID] = map[string]bool{}
		for _, attendee := range event.Attendees {
			eSet[event.ID][attendee.Email] = true
		}
	}

	return eSet, nil
}

//revive:disable

func (s *Server) ginRegisterAbsentees(c *gin.Context) {
	s.registerAbsentees()
}

func (s *Server) absenteeCronJob() {
	logrus.Info("Running absenteeCronJob")
	for i := 0; i < 10; i++ {
		doRetry := s.registerAbsentees()
		if !doRetry {
			logrus.Info("No absentee failures.")
			return
		}

		logrus.Info("One or more absentees failed, trying again in 60s.")
		time.Sleep(time.Second * 60)
	}
}

func (s *Server) startAbsenteeCronJob() error {
	c := cron.New()

	logrus.Infof("Starting absentee cron job with string '%s'", s.AbsenteeCronString)
	err := c.AddFunc(s.AbsenteeCronString, s.absenteeCronJob)
	if err != nil {
		return err
	}
	c.Start()
	logrus.Info("Successfully started cron job.")
	return nil
}

//revive:enable

//revive:disable-next-line:cyclomatic
func (s *Server) registerAbsentees() (anyFails bool) {

	logrus.Debug("Running registerAbsentees")

	//Get all events for last 24 hours
	absenteeSet, err := s.getTodaysEvents()
	if err != nil {
		logrus.Errorf("Failed to retrieve events for registerAbsentees: %s", err.Error())
		return
	}

	logrus.Debugf("Absenteeset size: %d", len(absenteeSet))

	//Register those who were actually present
	participationKeys, err := s.Redis.GetList("participation-*")
	if err != nil {
		logrus.Errorf("Failed to retrieve participation keys: %s", err.Error())
		return
	}
	logrus.Debugf("Participation keys length: %d", len(participationKeys))

	for _, key := range participationKeys {
		details := strings.Split(key, "-")
		activityID := sanitizeCalendarID(details[2])
		participantID := details[3]
		if absenteeSet[activityID] == nil {
			logrus.Warnf("Participation registered for event that did not exist today: %s", key)
			continue
		}

		absenteeSet[activityID][participantID] = false
		logrus.Debugf("%s was present at activity %s", participantID, activityID)
	}

	//Loop through participant set and register absentees
	for activityID, participants := range absenteeSet {
		absentList := []string{}

		//extract the absent
		for personEmail, absent := range participants {
			if absent {
				logrus.Debugf("%s was absent from activity %s",
					personEmail,
					activityID,
				)
				absentList = append(absentList, extractVismaID(personEmail))
			}
		}

		logrus.Debugf("Registering %d absentees from %s",
			len(absentList),
			activityID,
		)

		if len(absentList) > 0 {
			body := flyvo.RegisterAbsenceRequest{
				CourseID: strings.ToUpper(activityID[0:1]) + "." +
					strings.ToUpper(activityID[1:]),
				AbsenteeIds: absentList,
				AbsenceCode: "U",
			}

			gen := rpc.Generic{
				Path: rpc.PathRegisterAbsences,
			}

			gen.Body, err = json.Marshal(body)
			if err != nil {
				logrus.Fatalf("failed to marshal body: %s", err.Error())
			}

			//Inform Flyvo
			logrus.Debug("Awaiting clientside processing..")
			response, err := s.RPC.WaitForClientsideProcessing(&gen, time.Second*30)
			if err != nil {
				anyFails = true
				logrus.Errorf("Failed to register absentees due to error: %s", err.Error())
				continue
			} else if response.Status != http.StatusOK && response.Status != http.StatusNoContent {
				anyFails = true
				logrus.Errorf("Failed to register absentees, unexpected response (%d): %s",
					response.Status,
					response.Body,
				)
				continue
			}

			//On success, delete the keys
			logrus.Infof("Successfully registered absentees for activity '%s'", activityID)
		}
		err = s.Redis.DeleteRegex("participation-" +
			strings.ToUpper(activityID[0:1]) + "." +
			strings.ToUpper(activityID[1:]) + "*")
		if err != nil {
			logrus.Infof("Failed to delete participation  details for activity '%s'", activityID)
		}

		err = s.Redis.DeleteRegex("activity-" +
			strings.ToUpper(activityID[0:1]) + "." +
			strings.ToUpper(activityID[1:]) + "*")
		if err != nil {
			logrus.Infof("Failed to delete redis activity for activity '%s'", activityID)
		}
	}

	logrus.Debug("Done removing attendees")
	return
}
