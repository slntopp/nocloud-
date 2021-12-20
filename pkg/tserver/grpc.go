package tserver

import (
	"context"
	"errors"
	"io"
	"sync"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"go.uber.org/zap"
)



type tunnelHost struct {
	ctx    context.Context
	stream pb.SocketConnection_InitConnectionServer
}

type TunnelServer struct {
	mutex sync.Mutex
	pb.UnimplementedSocketConnectionServer
	fingerprints_hosts map[string]string
	hosts              map[string]tunnelHost
	request_id         map[uint32]chan ([]byte)

	col driver.Collection

	log *zap.Logger
}

const HOSTS_COLLECTION = "TunnelHosts"

func EnsureCollectionExists(logger *zap.Logger, db driver.Database) {
	log := logger.Named("EnsureCollectionExists")

	options := &driver.CreateCollectionOptions{
		KeyOptions: &driver.CollectionKeyOptions{AllowUserKeys: true, Type: "uuid"},
	}
	log.Debug("Checking Collection existence", zap.String("collection", HOSTS_COLLECTION))
	exists, err := db.CollectionExists(context.TODO(), HOSTS_COLLECTION)
	if err != nil {
		log.Fatal("Failed to check collection exists", zap.Error(err))
	}
	if exists {
		return
	}

	log.Debug("Collection doesn't exist, creating")
	_, err = db.CreateCollection(context.TODO(), HOSTS_COLLECTION, options)
	if err != nil {
		log.Fatal("Error creating Collection", zap.String("collection", HOSTS_COLLECTION), zap.Any("options", options), zap.Error(err))
	}
	log.Debug("Collection existence ensured")
}

func NewTunnelServer(log *zap.Logger, db driver.Database) *TunnelServer {
	col, _ := db.Collection(context.TODO(), HOSTS_COLLECTION)
	return &TunnelServer{
		fingerprints_hosts: make(map[string]string),
		hosts:              make(map[string]tunnelHost),
		request_id:         make(map[uint32]chan ([]byte)),

		col: col,

		log: log.Named("TunnelServer"),
	}
}

//Send data to client by grpc
func (s *TunnelServer) ScalarSendData(context.Context, *pb.HttpReQuest2Loc) (*pb.HttpReSp4Loc, error) {
	return &pb.HttpReSp4Loc{}, nil
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
		log.Error("Strange clienf sert", zap.String("Fingerprint", hex_sert_raw), zap.String("CommonName", cn))
		return errors.New("strange clienf sert:" + cn)
	}

	log.Info("Client connected", zap.String("Host", host))

	ctx, cancel := context.WithCancel(context.Background())

	s.mutex.Lock()
	s.hosts[host] = tunnelHost{ctx, stream}
	s.mutex.Unlock()
	defer func() {
		s.mutex.Lock()
		delete(s.hosts, host)
		s.mutex.Unlock()
	}()

	for {
		//receive json from location
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Error("Stream connection lost", zap.String("Host", host))
			} else {
				log.Error("stream.Recv", zap.String("Host", host))
			}
			cancel()
			return err
		}

		rid, ok := s.request_id[in.Id]
		if !ok {
			log.Error("Request does not exist", zap.String("Host", host))
			continue
		}
		rid <- in.Json

		log.Info("Received data from:", zap.String("Host", host))
	}
}
