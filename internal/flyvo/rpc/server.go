package rpc

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tktip/flyvo-api/internal/redis"
	"github.com/tktip/flyvo-api/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultPort = "50051"
)

//Server - the rpc server object
type Server struct {
	Port       string           `yaml:"port"`
	CertFile   string           `yaml:"cert"`
	KeyFile    string           `yaml:"key"`
	Redis      *redis.Connector `yaml:"redis"`
	Gcal       string           `yaml:"gcalUrl"`
	done       bool
	wg         sync.WaitGroup
	grpcServer *grpc.Server
	opts       []grpc.ServerOption

	asyncs asyncAsSync
}

//WaitForClientsideProcessing - Publish a request and wait for a response from the client
//if timeout duration is passed, the request is canceled and an errorAndClose is returned.
func (srv *Server) WaitForClientsideProcessing(
	g *rpc.Generic,
	timeout time.Duration,
) (
	rpc.Generic,
	error,
) {
	srv.asyncs.lock.Lock()
	reader := &genericReaderWriter{
		id:      uuid.New().String(),
		result:  make(chan *rpc.Generic),
		err:     make(chan error),
		generic: g,
	}
	srv.asyncs.readers = append(srv.asyncs.readers, reader)
	srv.asyncs.lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logrus.Debugf("Awaiting response from rpc client")
	generic, err := reader.Read(ctx)
	if err != nil {
		logrus.Debugf("Client side error: %s", err.Error())
		return rpc.Generic{}, err
	}

	logrus.Debugf("Successfully retreived client side data")
	return *generic, nil
}

func (srv *Server) init() (err error) {
	if srv.Port == "" {
		logrus.Warnf("No port provided, defaulting to '%s'", defaultPort)
		srv.Port = defaultPort
	}

	if srv.KeyFile == "" && srv.CertFile == srv.KeyFile {
		logrus.Info("No Cert/Key details. Running without certificate.")
		return
	} else if srv.CertFile == "" || srv.KeyFile == "" {
		return errors.New("missing cert or key file")
	}

	var creds credentials.TransportCredentials

	// Create the TLS credentials
	creds, err = credentials.NewServerTLSFromFile(srv.CertFile, srv.KeyFile)
	if creds != nil {
		srv.opts = append(srv.opts, grpc.Creds(creds))
		logrus.Info("Running with creds")
	}

	return
}

//listen waits for a connection from a flyvo-rpc-client
func (srv *Server) listen() {
	for !srv.done {
		srv.wg.Add(1)
		lis, err := net.Listen("tcp", ":"+srv.Port)
		if err != nil {
			logrus.Fatalf("failed to listen: %v", err)
		}

		srv.grpcServer = grpc.NewServer(srv.opts...)
		logrus.Infof("Listening for connections on: '%s'", lis.Addr().String())
		rpc.RegisterTipFlyvoServer(srv.grpcServer, srv)
		if err := srv.grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("failed to serve: %v", err)
		}
		srv.wg.Done()
	}
}

//Run - starts the web api
func (srv *Server) Run(ctx context.Context) {

	err := srv.init()
	if err != nil {
		logrus.Fatalf("Failed to initialize: %s", err.Error())
	}
	go srv.listen()

	select {
	case <-ctx.Done():
		srv.done = true
		logrus.Infof("Received ctx.Done, stopping gracefully.")
		srv.grpcServer.GracefulStop()
		srv.wg.Wait()
		logrus.Info("Stopped gracefully. Shutting down.")
		os.Exit(0)
	}
}
