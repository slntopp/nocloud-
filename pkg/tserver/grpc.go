package tserver

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"go.uber.org/zap"
)

//Keeps all connections from clients (hosts) and ctx.WithCancel cut all http-requests if connection lost
type tunnelHost struct {
	ctx       context.Context
	stream    pb.SocketConnection_InitConnectionServer
	resetConn chan (error)
}

//struct for GRPC interface SocketConnectionServer and exchange data between methods
type TunnelServer struct {
	mutex sync.Mutex
	pb.UnimplementedSocketConnectionServer
	fingerprints_hosts map[string]string
	hosts              map[string]tunnelHost
	request_id         map[uint32]chan ([]byte)

	col driver.Collection

	log *zap.Logger
}

//Initialize new struct for GRPC interface
func NewTunnelServer(log *zap.Logger, db driver.Database) *TunnelServer {
	col, _ := db.Collection(context.TODO(), HOSTS_COLLECTION)
	est := true //todo &true not work
	col.EnsureFullTextIndex(context.TODO(), []string{"host"}, &driver.EnsureFullTextIndexOptions{
		MinLength:    2,
		InBackground: true,
		Name:         "textIndex",
		Estimates:    &est,
	})

	return &TunnelServer{
		fingerprints_hosts: make(map[string]string),
		hosts:              make(map[string]tunnelHost),
		request_id:         make(map[uint32]chan ([]byte)),

		col: col,

		log: log.Named("TunnelServer"),
	}
}

//Reset connections if his fp/hosts edit/remove in db
func (s *TunnelServer) resetConn(in *pb.HostFingerprint) {
	log := s.log.Named("Edition in DB")

	_, ok := s.hosts[in.Host]
	if ok {
		s.mutex.Lock()
		s.hosts[in.Host].resetConn <- nil
		s.mutex.Unlock()

		log.Info("Connection closed by DB edition", zap.String("Host", in.Host))
	}

}

func (s *TunnelServer) WaitForConnection(host string) error {
	r := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(time.Second)
			_, ok := s.hosts[host]
			if ok {
				r <- true
			}
		}
	}()
	select {
    case <- r:
        return nil
    case <- time.After(30 * time.Second):
        return errors.New("Connection timeout exceeded")
    }
}

//Initiate soket connection from Location
func (s *TunnelServer) InitConnection(stream pb.SocketConnection_InitConnectionServer) error {
	log := s.log.Named("InitConnection")
	peer, _ := peer.FromContext(stream.Context())
	raw := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Raw
	hex_sert_raw := MakeFingerprint(raw)

	host, ok := s.fingerprints_hosts[hex_sert_raw]
	if !ok {
		cn := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Subject.CommonName
		log.Error("Unregistered client Certificate", zap.String("Fingerprint", hex_sert_raw), zap.String("CommonName", cn))
		return errors.New("Unregistered Certificate: " + cn)
	}

	log.Info("Client connected", zap.String("Host", host))

	ctx, cancel := context.WithCancel(context.Background())

	s.mutex.Lock()
	s.hosts[host] = tunnelHost{ctx, stream, make(chan error)}
	s.mutex.Unlock()

	defer func() {
		cancel()
		s.mutex.Lock()
		delete(s.hosts, host)
		s.mutex.Unlock()
	}()

	go func() {
		var errtmp error
		for {
			//receive json from location
			in, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.Error("Stream connection lost", zap.String("Host", host))
				} else {
					log.Error("stream.Recv", zap.String("Host", host), zap.Error(err))
				}
				// s.hosts[host].resetConn <- err //if client disconnected: panic: runtime error: invalid memory address or nil pointer dereference
				errtmp = err
				break
			}

			rid, ok := s.request_id[in.Id]
			if !ok {
				log.Error("Request does not exist", zap.String("Host", host))
				continue
			}
			rid <- in.Json

			log.Info("Received data from:", zap.String("Host", host))
		}

		s.hosts[host].resetConn <- errtmp
	}()

	err1 := <-s.hosts[host].resetConn
	log.Info("Connection closed", zap.String("Host", host), zap.Error(err1))
	return err1
}
