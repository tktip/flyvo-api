package httperr

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// Error represents an error being able to write to an http.ResponseWrite and create logrus
// entries
type Error interface {
	error
	HTTPWrite(w http.ResponseWriter)
	Entry() *logrus.Entry
	Warn()
	Info()
	WithCall(callContext string) Error
	WithField(name string, value interface{}) Error
}

// Error wraps a native error with an http error
type httpErr struct {
	// status is the code that should be returned to the requester
	status int
	// err is the error that occurred. This is returned to the user if Text is empty
	err error
	// text is the text returned to the user. Err.Error() is used if this is empty
	text string
	// calls contains a string representing each call "down the stack" until the error occurred
	calls []string
	// fields is a map of fields added to the logrus entry
	fields logrus.Fields
}

func (err httpErr) Error() string {
	return strings.Join(append(err.calls, err.err.Error()), ": ")
}

// HTTPWrite writes this error's status and text (err.Error() if text is empty) to the given http
// response writer
func (err httpErr) HTTPWrite(w http.ResponseWriter) {
	w.WriteHeader(err.status)
	if err.text != "" {
		w.Write([]byte(err.text))
		return
	}
	w.Write([]byte(err.Error()))
}

// HTTPWrite writes this error's status and text (err.Error() if text is empty) to the given http
// response writer
func (err httpErr) HTTPWriteWithID(w http.ResponseWriter, id string) {
	w.Header().Add("X-Tip-Correlation-Id", id)
	err.HTTPWrite(w)
}

// Entry creates a logrus entry with the http status as an entry-field
func (err httpErr) Entry() *logrus.Entry {
	return logrus.WithFields(err.fields).WithField("status", err.status)
}

// Warn logs this error with the warn log-level
func (err httpErr) Warn() {
	err.Entry().Warn(err)
}

// Info logs this error with the Info log-level
func (err httpErr) Info() {
	err.Entry().Info(err)
}

// WithCall creates a copy of this error with the given call prepended to the list of calls.
// The list of calls is joined with ": " and prepended on to the error-message when Error() is
// called like so:
// call3: call2: call1: Error()-message here
func (err httpErr) WithCall(call string) Error {
	err.calls = append([]string{call}, err.calls...)
	return err
}

func (err httpErr) WithField(name string, value interface{}) Error {
	err.fields[name] = value
	return err
}

// New creates a new error with the given err and status and an empty text
func New(status int, err error) Error {
	return httpErr{status, err, "", []string{}, make(logrus.Fields)}
}

// Newf creates a new error with the given err and status and an empty text
func Newf(status int, str string, args ...interface{}) Error {
	return httpErr{
		status: status,
		err:    fmt.Errorf(str, args...),
		fields: make(logrus.Fields),
	}
}

// NewText creates a new error with the given err, status and text
func NewText(status int, err error, text string) Error {
	return httpErr{
		status: status,
		err:    err,
		text:   text,
		fields: make(logrus.Fields),
	}
}
