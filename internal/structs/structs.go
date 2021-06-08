package structs

import "time"

//EventSickLeave - register sick leave request
type EventSickLeave struct {
	AbsenceCode string     `json:"absenceCode"`
	Start       *time.Time `json:"start"`
	End         *time.Time `json:"end"`
}

//RegisterParticipation - register participation request struct
type RegisterParticipation struct {
	ActivityIDExternal string `json:"activityId"`
	ParticipantID      string `json:"participantId"`
}
