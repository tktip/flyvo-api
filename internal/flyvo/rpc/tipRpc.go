package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/pkg/rpc"
	"github.com/tktip/google-calendar/pkg/googlecal"
)

var (
	invalidChars = regexp.MustCompile(`([a-z0-9])*`)
)

const (
	addEvent          = "event/create"
	updateEvent       = "event/update"
	deleteEvent       = "event/delete/"
	removeParticipant = "event/participants/"
)

func sanitizeCalendarID(ID string) string {
	return strings.Join(invalidChars.FindAllString(strings.ToLower(ID), -1), "")
}

//revive:disable:cyclomatic
func (srv *Server) sendToGcal(
	ctx context.Context,
	event googlecal.Event,
	method string,
	endpoint string,
) (
	*rpc.Generic,
	error,
) {

	if event.ID != nil {
		*event.ID = sanitizeCalendarID(*event.ID)
		{
			if event.End != nil && strings.HasSuffix(*event.End, "Z") {
				*event.End = correctTime(*event.End)
			}
			if event.Start != nil && strings.HasSuffix(*event.Start, "Z") {
				*event.Start = correctTime(*event.Start)
			}
		}
	}

	body, err := json.Marshal(event)
	if err != nil {
		logrus.Errorf("Failed to marshal event: %s", err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, srv.Gcal+endpoint, bytes.NewReader(body))
	query := req.URL.Query()
	query.Set("broadcastChanges", "false")
	query.Set("guestsVisible", "false")
	query.Set("guestsMayInvite", "false")
	query.Set("guestsCanModify", "false")
	query.Set("guestsAutoAccept", "true")
	query.Set("privateEvent", "true")

	req.URL.RawQuery = query.Encode()
	if err != nil {
		logrus.Errorf("Failed to generate 'create event' request: %s", err.Error())
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Failed to perform event request: %s", err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected status [%s]: %s", resp.Status, body)
		logrus.Errorf("Failed to create event: %s", err.Error())
		return nil, err
	}

	return &rpc.Generic{
		Body:   []byte(fmt.Sprintf(`{"eventId":"%s"}`, body)),
		Status: http.StatusOK,
	}, nil
}

func getParticipantsAsMails(participants []*rpc.Participant) []string {
	mails := []string{}
	for _, participant := range participants {
		//trovo?
		mail := strings.ToLower(fmt.Sprintf("%s%s%s@trovo.no",
			participant.GivenName[:1],
			participant.Surname[:1],
			participant.VismaId))
		mails = append(mails, mail)
	}
	return mails
}

// PublishEvent publishes event to google.
func (srv *Server) PublishEvent(ctx context.Context, in *rpc.Event) (*rpc.Generic, error) {
	logrus.Debugf("Received publishEvent: %+v", *in)
	mails := getParticipantsAsMails(in.Participants)
	gEvent := googlecal.Event{
		ID:           &in.VismaActivityId,
		Title:        &in.ActivityTitle,
		Location:     &in.Location, //Consider in.Room
		Start:        &in.From,
		End:          &in.To,
		Description:  &in.ActivityTitle,
		Participants: &mails,
	}

	return srv.sendToGcal(ctx, gEvent, http.MethodPost, addEvent)
}

// UpdateEvent updates event in google
func (srv *Server) UpdateEvent(ctx context.Context, in *rpc.Event) (*rpc.Generic, error) {

	logrus.Debugf("Received updateEvent: %+v", *in)

	mails := getParticipantsAsMails(in.Participants)

	gEvent := googlecal.Event{
		ID:           &in.VismaActivityId,
		Title:        &in.ActivityTitle,
		Location:     &in.Location, //Consider in.Room
		Start:        &in.From,
		End:          &in.To,
		Description:  &in.ActivityTitle,
		Participants: &mails,
	}

	return srv.sendToGcal(ctx, gEvent, http.MethodPut, updateEvent)
}

// DeleteEvent performs event delete in google.
func (srv *Server) DeleteEvent(ctx context.Context, in *rpc.String) (*rpc.Generic, error) {
	logrus.Debugf("Received deleteEvent: %v", in.Value)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		srv.Gcal+deleteEvent+sanitizeCalendarID(in.Value),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status '%s': %s", resp.Status, body)
	}

	return &rpc.Generic{
		Body:   []byte(`ok`),
		Status: http.StatusOK,
	}, nil
}

// RemoveFromEvent performs event delete in google.
func (srv *Server) RemoveFromEvent(ctx context.Context, in *rpc.String) (*rpc.Generic, error) {
	logrus.Debugf("Received removeFromEvent: %v", in.Value)

	//Validate input data
	data := strings.Split(in.Value, "/")
	if len(data) != 2 || data[0] == "" || data[1] == "" {
		return nil, errors.New("bad data, format is [event/participant]")
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		srv.Gcal+removeParticipant+in.Value,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status '%s': %s", resp.Status, body)
	}

	return &rpc.Generic{
		Body:   []byte(`ok`),
		Status: http.StatusOK,
	}, nil
}

// HandleGeneric responds to generic requests
func (srv *Server) HandleGeneric(ctx context.Context, in *rpc.Generic) (*rpc.Generic, error) {
	logrus.Debugf("Received generic: %+v", in)
	if in.Path == ping {
		return srv.ping(ctx, *in), nil
	}

	return &rpc.Generic{
		Status: http.StatusBadRequest,
		Body:   []byte("unknown path"),
	}, errors.New("unknown path")
}

// ProcessRequests processes requests from web clients.
// I.e. the rpc server stocks up web client requests, and the rpc client retrieves and
// processes them, and the response from the rpc client is proxied back to web client.
// NOTE: This function is called remotely from the RPC client.
func (srv *Server) ProcessRequests(client rpc.TipFlyvo_ProcessRequestsServer) error {
	logrus.Debugf("Locking asyncs lock")

	//Retrieve any requests received from frontend
	srv.asyncs.lock.Lock()

	//Just stop on no requests to process.
	if len(srv.asyncs.readers) == 0 {
		logrus.Debug("No new requests from users")
		srv.asyncs.lock.Unlock()
		return nil
	}

	//Grab readwriters
	writers := map[string]*genericReaderWriter{}
	for i := range srv.asyncs.readers {
		writers[srv.asyncs.readers[i].id] = srv.asyncs.readers[i]
	}
	srv.asyncs.readers = []*genericReaderWriter{}
	logrus.Debugf("Unlocking asyncs lock")
	srv.asyncs.lock.Unlock()

	connClosed := false

	logrus.Debugf("Processing %d end user requests", len(writers))

	//Process requests
	for reqID, writer := range writers {
		logrus.Debugf("%s: Processing req.", reqID)

		//if connection is closed, inform all readers
		if connClosed {
			logrus.Debugf("%s: Connection was closed.", reqID)
			writer.errorAndClose(io.EOF)
			continue
		}

		//Send request from api
		logrus.Debugf("%s: Sending request to client", reqID)
		err := client.Send(writer.generic)
		if err != nil {
			logrus.Debugf("%s: Could not send request to client.", reqID)
			writer.errorAndClose(err)
			continue
		}

		logrus.Debugf("%s: Awaiting client response...", reqID)

		//Get response
		g, err := client.Recv()

		//connection closed on client side
		if err == io.EOF {
			logrus.Debugf("%s: Got io.EOF from client - seems connection is closed.", reqID)
			connClosed = true
		}

		if err != nil {
			logrus.Warnf("%s: Got error from client: %s", err.Error(), reqID)
			writer.errorAndClose(err)
		} else {
			logrus.Debugf("%s: Successfully got response from client.", reqID)
			writer.writeAndClose(g)
		}
		logrus.Debugf("%s: Done processing.", reqID)
	}

	logrus.Debugf("Done with client processing of end user reuqests.")

	return nil
}
