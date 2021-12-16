package tserver

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/slntopp/nocloud-tunnel-mesh/pkg/logger"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	lg *zap.Logger
)

func init() {
	lg = logger.NewLogger()

	viper.AutomaticEnv()
	viper.SetDefault("GRPC_PORT", "8080")
}

type tunnelHost struct {
	ctx    context.Context
	stream pb.SocketConnection_InitConnectionServer
}

type tunnelServer struct {
	mutex sync.Mutex
	pb.UnimplementedSocketConnectionServer
	fingerprints_hosts map[string]string
	hosts              map[string]tunnelHost
	request_id         map[uint32]chan ([]byte)
}

//Send data to client by grpc
func (s *tunnelServer) ScalarSendData(context.Context, *pb.HttpReQuest2Loc) (*pb.HttpReSp4Loc, error) {
	return &pb.HttpReSp4Loc{}, nil
}

//Initiate soket connection from Location
func (s *tunnelServer) InitConnection(stream pb.SocketConnection_InitConnectionServer) error {

	peer, _ := peer.FromContext(stream.Context())
	raw := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Raw
	hex_sert_raw := getFingerprint(raw)

	host, ok := s.fingerprints_hosts[hex_sert_raw]
	if !ok {
		cn := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Subject.CommonName
		lg.Error("Strange clienf sert", zap.String("Fingerprint", hex_sert_raw), zap.String("CommonName", cn))
		return errors.New("strange clienf sert:" + cn)
	}

	lg.Info("Client connected", zap.String("Host", host))

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
				lg.Error("Stream connection lost", zap.String("Host", host))
			} else {
				lg.Error("stream.Recv", zap.String("Host", host))
			}
			cancel()
			return err
		}

		rid, ok := s.request_id[in.Id]
		if !ok {
			lg.Error("Request does not exist", zap.String("Host", host))
			continue
		}
		rid <- in.Json

		lg.Info("Received data from:", zap.String("Host", host))

	}
}
//Start GRPC Server, start HttpServer, initialize DB fingerprints
func StartGRPCServer() {
	port := viper.GetString("GRPC_PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.Fatal("failed to listen:", zap.Error(err))
	}

	var opts []grpc.ServerOption

	//openssl req -new -newkey rsa:4096 -x509 -sha256 -days 30 -nodes -out server.crt -keyout server.key
	cert, err := tls.LoadX509KeyPair("/cert/server.crt", "/cert/server.key")
	if err != nil {
		lg.Fatal("server: loadkeys:", zap.Error(err))
	}

	// Note if we don't tls.RequireAnyClientCert client side certs are ignored.
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// ClientCAs:    caCertPool,//It work! Peer client sertificates autenification
		ClientAuth: tls.RequireAnyClientCert,
		// VerifyPeerCertificate: is called only after normal certificate verification https://pkg.go.dev/crypto/tls#Config
		// InsecureSkipVerify: false,
		InsecureSkipVerify: true,
	}
	cred := credentials.NewTLS(config)

	opts = append(opts, grpc.Creds(cred))
	
	grpcServer := grpc.NewServer(opts...)

	newServiceImpl := &tunnelServer{
		fingerprints_hosts: make(map[string]string),
		hosts:              make(map[string]tunnelHost),
		request_id:         make(map[uint32]chan ([]byte)),
	}

	newServiceImpl.getHostFingerprintsFromDB()

	pb.RegisterSocketConnectionServer(grpcServer, newServiceImpl)

	srv := newServiceImpl.startHttpServer()

	lg.Info("gRPC-Server Listening on 0.0.0.0:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		lg.Fatal("failed to serve grpc:", zap.Error(err))
	}

	srv.Shutdown(context.TODO())
}
