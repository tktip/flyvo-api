package flyvo

import (
	"time"
)

//GetCoursesResponse - response on get courses.
type GetCoursesResponse []VismaCourse

//VismaCourse - course as sent by visma
type VismaCourse struct {
	VismaID string `json:"vismaActivityId"`
	From    string `json:"timeFrom"` //aa:bb
	To      string `json:"timeTo"`   //aa:bb
	Date    string `json:"date"`     //ddMMyyyy
	Place   string `json:"place"`
	Rom     string `json:"room"`
}

//GetCoursesRequest - accepted request on get courses
type GetCoursesRequest struct {
	//TeacherID string    `json:"teacherId"`
	FromDate time.Time `json:"fromDate"`
	ToDate   time.Time `json:"toDate"`
}

//RegisterAbsenceRequest - accepted request on register absence
type RegisterAbsenceRequest struct {
	CourseID    string   `json:"vismaActivityId"`
	AbsenceCode string   `json:"absenceCode"`
	AbsenteeIds []string `json:"absentees"`
}

//GetUnauthorizedAbsenceRequest - accepted request body on get absences
type GetUnauthorizedAbsenceRequest struct {
	VismaID  string    `json:"vismaId"`
	FromDate time.Time `json:"from"`
	ToDate   time.Time `json:"to"`
}

//UnauthorizedAbsenceActivity - response on unauthorized absence
type UnauthorizedAbsenceActivity struct {
	ActivityID           string `json:"vismaActivityId"`
	NumberOfInvalidHours string `json:"numberOfInvalidHours"`
}

//GetUnauthorizedAbsenceResponse - response get absences
type GetUnauthorizedAbsenceResponse struct {
	VismaID    string                        `json:"vismaId"`
	GivenName  string                        `json:"givenName"`
	Surname    string                        `json:"surname"`
	Activities []UnauthorizedAbsenceActivity `json:"activities"`
}

//GetSickLeavesRequest - accepted request get sick leave
type GetSickLeavesRequest struct {
	VismaID string `json:"vismaId"`
	ToDate  string `json:"toDate"`
}

//GetSickLeavesResponse response struct on sick leave
type GetSickLeavesResponse struct {
	VismaID        string `json:"vismaId"`
	GivenName      string `json:"givenName"`
	Surname        string `json:"surname"`
	SickLeaveCount int    `json:"numSelfCertifications"`
	SickChildCount int    `json:"sumSelfCertificationsChildren"`
}

//RegisterSickLeave struct for regeistering sick leave
type RegisterSickLeave struct {
	VismaID  string `json:"vismaId"`
	Code     string `json:"absenceCode"`
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
}
