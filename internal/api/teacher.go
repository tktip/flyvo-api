package api

import "github.com/gin-gonic/gin"

//revive:disable:unused-receiver

func (s *Server) getIsTeacher(c *gin.Context) {
	if !isTeacherWithoutError(c) {
		c.Writer.WriteHeader(204)
		return
	}
	c.Writer.WriteHeader(200)
}
