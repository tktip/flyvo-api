package googlecal

import "google.golang.org/api/calendar/v3"

//Event contains event data
type Event struct {

	//pointers to allow empty values
	ID           *string                  `json:"id"`
	Title        *string                  `json:"title"`
	Description  *string                  `json:"description"`
	Location     *string                  `json:"location"`
	Start        *string                  `json:"startDateTime"`
	End          *string                  `json:"endDateTime"`
	Participants *[]string                `json:"participants"`
	Organizer    *calendar.EventOrganizer `json:"organizer"`
}
