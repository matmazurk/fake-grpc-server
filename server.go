package fakegrpcserver

import (
	"context"
	"log"
	"net"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type fakeGRPCServer struct {
	lis *bufconn.Listener

	serv *grpc.Server
}

type Register func(*grpc.Server)

func NewFakeServer(register Register) *fakeGRPCServer {
	s := grpc.NewServer()
	register(s)

	return &fakeGRPCServer{
		serv: s,
	}
}

func (s *fakeGRPCServer) Start() func() {
	s.lis = bufconn.Listen(bufSize)

	go func() {
		if err := s.serv.Serve(s.lis); err != nil {
			log.Fatalf("fake GRPC server exited with error: %v", err)
		}
	}()

	return func() {
		s.lis = nil
		s.serv.GracefulStop()
	}
}

func (m *fakeGRPCServer) Conn() (*grpc.ClientConn, error) {
	if m.lis == nil {
		return nil, errors.New("server must be started")
	}

	conn, err := grpc.Dial(
		"bufnet",
		grpc.WithContextDialer(m.bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial bufnet")
	}

	return conn, nil
}

func (m *fakeGRPCServer) bufDialer(context.Context, string) (net.Conn, error) {
	if m.lis == nil {
		return nil, errors.New("cannot dial on nil listener")
	}
	return m.lis.Dial()
}
