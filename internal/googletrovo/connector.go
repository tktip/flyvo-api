package googletrovo

import (
	"context"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

//Connector - provides connection to google
type Connector struct {
	Creds        string `yaml:"creds"`
	AdminUser    string `yaml:"adminUser"`
	TeacherGroup string `yaml:"teacherGroup"`
	service      *admin.MembersService
}

func (c *Connector) getMemberService() *admin.MembersService {
	if c.service != nil {
		return c.service
	}

	config, err := google.JWTConfigFromJSON([]byte(c.Creds),
		admin.AdminDirectoryGroupMemberReadonlyScope,
		admin.AdminDirectoryGroupReadonlyScope,
	)

	config.Subject = c.AdminUser

	service, err := admin.NewService(
		context.Background(), option.WithHTTPClient(config.Client(context.Background())))
	if err != nil {
		log.Fatalf("Could not create directory service client => {%s}", err)
	}

	c.service = admin.NewMembersService(service)
	return c.service
}

//IsMemberOfTeacherGroup - checks whether user is member of specified group
func (c *Connector) IsMemberOfTeacherGroup(email string) (bool, error) {
	l := c.getMemberService().HasMember(c.TeacherGroup, email)
	m, err := l.Do()
	if err != nil {
		gerr, isGoogleAPIErr := err.(*googleapi.Error)
		if isGoogleAPIErr &&
			strings.Contains(gerr.Message, "memberKey") {

			logrus.Errorf("User seems to not exist in google: %s", email)
			return false, nil
		}
		return false, err
	}
	return m.IsMember, nil
}
