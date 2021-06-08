package rpc

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/pkg/rpc"
)

//genericReaderWriter is used to maintain a channel from frontend request
//to flyvo client. An channel stays open until closed, either because it is
//filled, or because the error channel is filled. The frontend api will then
//grab the contents of the respective channel and respond to end user.
//It looks like this:
//	[web] <-> [web-api] <-genericReaderWriter-> [rpc-server] <-> [flyvo-rpc-client]
type genericReaderWriter struct {
	sync.Mutex
	closer  sync.Once
	closed  bool
	id      string
	result  chan *rpc.Generic
	err     chan error
	generic *rpc.Generic
}

//asyncAsSync is a set of open requests from frontend. I.e. frontend functions
//add their requests to this set, and then await responses put in those requests'
//separate writers.
type asyncAsSync struct {
	lock    sync.Mutex
	readers []*genericReaderWriter
}

//This function writes data to the error channel and then closes the writer,
//rejecting further writes.
func (g *genericReaderWriter) errorAndClose(err error) {
	logrus.Debugf("errorAndClose lock waiting with error: %s", err.Error())
	g.Lock()
	logrus.Debugf("errorAndClose locked with error: %s", err.Error())
	if !g.closed {
		g.err <- err
		g.close()
	} else {
		logrus.Warn("Write to closed err chan")
	}
	g.Unlock()
}

//This function writes data to the result channel and then closes the writer,
//rejecting further writes.
func (g *genericReaderWriter) writeAndClose(generic *rpc.Generic) {
	logrus.Debugf("writeAndClose (%s) lock waiting with msg: %s",
		generic.Path,
		generic.Body,
	)
	g.Lock()
	logrus.Debugf("writeAndClose (%s) locked with msg: %s",
		generic.Path,
		generic.Body,
	)
	if !g.closed {
		g.result <- generic
		g.close()
	} else {
		logrus.Warn("Write to closed gen chan")
	}
	g.Unlock()
}

func (g *genericReaderWriter) close() {
	logrus.Debug("Readwriter close called")
	if g.closed {
		logrus.Warn("Close called after already closed")
	}

	g.closer.Do(func() {
		close(g.err)
		close(g.result)
		g.closed = true
	})
}

//Read waits for either a response on channels, or for the context to complete.
func (g *genericReaderWriter) Read(ctx context.Context) (*rpc.Generic, error) {
	select {
	case err := <-g.err:
		return nil, err
	case res := <-g.result:
		return res, nil
	case <-ctx.Done():
		g.Lock()
		g.close()
		g.Unlock()
		return nil, context.DeadlineExceeded
	}
}
