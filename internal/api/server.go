package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/internal/flyvo/rpc"
	"github.com/tktip/flyvo-api/internal/googletrovo"
	"github.com/tktip/flyvo-api/internal/redis"
	"github.com/tktip/flyvo-api/pkg/swagex"
	//	"github.com/sirupsen/logrus"
)

const (
	eventList = "event/list/%s/%s"
	eventGet  = "event/get/"
)

//Server - api server object
type Server struct {
	Port  string                `yaml:"port"`
	Debug bool                  `yaml:"debug"`
	Redis redis.Connector       `yaml:"redis"`
	RPC   rpc.Server            `yaml:"rpc"`
	Trovo googletrovo.Connector `yaml:"trovo"`

	GcalURL string `yaml:"gcalUrl"`
	QrURL   string `yaml:"qrUrl"`

	ParticipantURL string `yaml:"participantUrl"`

	AbsenteeCronString string `yaml:"absentCron"`
}

//Run starts the api
func (s *Server) Run() error {

	ctx, cancel := context.WithCancel(context.Background())
	s.RPC.Redis = &s.Redis

	defer cancel()

	go s.RPC.Run(ctx)

	if s.AbsenteeCronString != "" {
		logrus.Infof("Starting absentee cron job (%s)", s.AbsenteeCronString)
		err := s.startAbsenteeCronJob()
		if err != nil {
			return err
		}
	} else {
		logrus.Warn("No cron string provided")
	}

	//Starting Gin
	r := gin.New()
	r.Use(gin.Logger()) // request logging

	r.Use(s.extractPersonFromCookie)
	r.Use(s.setIsTeacher)

	r.GET("/generate/qr", s.generateQrCode)
	r.GET("/generate/participationId", s.generateParticipationID)

	r.POST("/absence/registerSickLeave", s.registerSickleave)
	r.GET("/absence/getSickleaves/:to", s.getSickleaves)
	r.GET("/absence/count/:from/:to", s.getAbsenceCount)
	r.GET("/event/retrieve/:from/:to", s.getEventsForTeacher)
	r.GET("/event/participate", s.registerParticipation)
	r.GET("/isTeacher", s.getIsTeacher)

	r.GET("/api-doc", swagex.SwaggerEndpoint)

	p := ":8080"
	if s.Port != "" {
		p = ":" + s.Port
	}
	return r.Run(p)
}
