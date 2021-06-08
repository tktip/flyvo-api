package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	jwtsessions "github.com/tktip/google-auth-proxy/pkg/jwt-sessions"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//revive:disable:unused-receiver

var (
	letterRunes    = []rune("1234567890")
	teacherTimeout = time.Minute * 60
)

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//revive:disable-next-line:cyclomatic
func (s *Server) mailBelongsToTeacher(email string) (isTeacher bool, err error) {
	if email == "" {
		return false, errors.New("no email")
	}

	var cachedVal string
	cachedVal, err = s.Redis.GetStringValue("teacher-" + email)
	if err == redis.Nil || cachedVal == "false" {
		return false, nil
	} else if err != nil {
		return false, err
	} else if cachedVal == "true" {
		return true, nil
	}

	isTeacher, err = s.Trovo.IsMemberOfTeacherGroup(email)
	if err != nil {
		return
	}

	val := "true"
	if !isTeacher {
		val = "false"
	}

	//Disregarding this error as it won't cause any issues.
	err = s.Redis.WriteValue("teacher-"+email, val, teacherTimeout)
	if err != nil {
		logrus.Errorf("Failed to write value to redis: %s", err.Error())
	}

	return isTeacher, nil

}

func (s *Server) extractPersonFromCookie(c *gin.Context) {
	if true {
		c.Set("person", &jwtsessions.GToken{
			FamilyName: "Bentdal",
			GivenName:  "Simen",
			Email:      "simen.bentdal@trondheim.kommune.no",
		})
		return
	}
	header := c.GetHeader("userInfo")

	if header == "" {
		logrus.Info("Missing userinfo. Request did most likely not from auth-proxy")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":     "missing oauth header",
			"errorCode": CodeNotAuthenticated,
		})
		return
	}

	person := &jwtsessions.GToken{}
	err := json.Unmarshal([]byte(header), person)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"errorCode": CodeInternalErrorGeneral,
		})
		return
	}
	c.Set("person", person)
	c.Next()
}

func (s *Server) setIsTeacher(c *gin.Context) {
	if true {
		c.Set("teacher", true)
		return
	}
	person, ok := c.Get("person")
	if !ok {
		logrus.Error("Reached setIsTeacher without person set.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":     "invalid internal state",
			"errorCode": CodeInternalErrorGeneral,
		})
		return
	}

	p, ok := person.(*jwtsessions.GToken)
	if !ok {
		fmt.Print()
	}

	isTeacher, err := s.mailBelongsToTeacher(p.Email)
	if err != nil {

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"errorCode": CodeInternalErrorGeneral,
		})
		return
	}

	c.Set("teacher", isTeacher)
	c.Next()
}

func extractVismaID(email string) string {
	split := strings.Split(email, "@")
	//Example, for pål testesen: pt12345@...
	//vismaID = 12345.

	if len(split) == 0 || len(split[0]) < 3 {
		logrus.Warnf("Mail was bad: %s", email)
		return ""
	}
	return split[0][2:]
}

func isTeacherWithoutError(c *gin.Context) bool {
	t, ok := c.Get("teacher")
	if !ok {
		return false
	}
	if !t.(bool) {
		return false
	}
	return true
}

func isTeacher(c *gin.Context) bool {
	t, ok := c.Get("teacher")
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			codedErrorResponse(
				"invalid internal state",
				CodeInternalErrorGeneral,
			))
		return false
	}

	if !t.(bool) {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			codedErrorResponse(
				"must have teacher privileges",
				CodeForbidden,
			))
		return false
	}

	return true
}

func getPersonObject(c *gin.Context) (*jwtsessions.GToken, bool) {
	t, ok := c.Get("person")
	if !ok || t == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			codedErrorResponse(
				"invalid internal state",
				CodeInternalErrorGeneral,
			))
		return nil, false
	}

	return t.(*jwtsessions.GToken), true
}
