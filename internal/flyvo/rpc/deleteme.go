package rpc

import (
	"time"

	"github.com/sirupsen/logrus"
)

var (
	winter2020 = time.Date(2020, 10, 25, 0, 0, 0, 0, time.UTC)
	summer2021 = time.Date(2021, 3, 28, 0, 0, 0, 0, time.UTC)
	winter2021 = time.Date(2021, 10, 31, 0, 0, 0, 0, time.UTC)
	summer2022 = time.Date(2022, 03, 27, 0, 0, 0, 0, time.UTC)
)

func isSummer(t time.Time) bool {
	if t.After(summer2022) {
		return true
	} else if t.After(winter2021) {
		return false
	} else if t.After(summer2021) {
		return true
	} else if t.After(winter2020) {
		return false
	}

	return true
}

func correctTime(incorrect string) string {
	if incorrect == "" {
		return ""
	}

	//Panic?
	t, err := time.Parse(time.RFC3339, incorrect)
	if err != nil {
		logrus.Errorf("BAD TIME '%s': %s", incorrect, err.Error())
		return incorrect
	}

	if isSummer(t) {
		return (incorrect)[:len(incorrect)-1] + "+02:00"
	}
	return (incorrect)[:len(incorrect)-1] + "+01:00"
}
