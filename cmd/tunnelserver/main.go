package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

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
	viper.SetDefault("HTTP_PORT", "8090")
	viper.SetDefault("SECURE", true)
}

type tunnelServer struct {
	mutex sync.Mutex
	pb.UnimplementedSocketConnectionServer
	hosts_fingerprints map[string]string
	hosts_sokets       map[string]pb.SocketConnection_InitConnectionServer
	request_id         map[uint32]chan ([]byte)
}

func getFingerprint(c []byte) string {
	sum := sha256.Sum256(c)
	return hex.EncodeToString(sum[:])
}

func (s *tunnelServer) getHostFingerprintsFromDB() {
	fp := "419d3335b2b533526d4e7f6f1041b3c492d086cad0f5876739800ffd51659545"
	s.hosts_fingerprints[fp] = "httpbin.org"
	// s.hosts_fingerprints[fp] ="reqbin.com"
	// s.hosts_fingerprints[fp] ="zero.client.net"
	// s.hosts_fingerprints[fp] ="ione-cloud.net"
}

//Send data to client by grpc
func (s *tunnelServer) ScalarSendData(context.Context, *pb.HttpReQuest2Loc) (*pb.HttpReSp4Loc, error) {
	return &pb.HttpReSp4Loc{}, nil
}

//Add host/Fingerprint to DB
func (s *tunnelServer) Add(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {

	_, ok := s.hosts_fingerprints[in.Fingerprint]
	if ok {
		return &pb.HostFingerprintResp{Message: in.Host + " exists!", Sucsess: false}, nil
	}

	s.mutex.Lock()
	s.hosts_fingerprints[in.Fingerprint] = in.Host
	s.mutex.Unlock()

	return &pb.HostFingerprintResp{Message: in.Host + " added", Sucsess: true}, nil
}

//Edit host/Fingerprint to DB
func (s *tunnelServer) Edit(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {

	_, ok := s.hosts_fingerprints[in.Fingerprint]
	if !ok {
		return &pb.HostFingerprintResp{Message: in.Host + " not exist!", Sucsess: false}, nil
	}
	s.mutex.Lock()
	s.hosts_fingerprints[in.Fingerprint] = in.Host
	s.mutex.Unlock()

	return &pb.HostFingerprintResp{Message: in.Host + " edited", Sucsess: true}, nil
}

//Delete host/Fingerprint to DB
func (s *tunnelServer) Delete(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {
	s.hosts_fingerprints[in.Fingerprint] = in.Host

	_, ok := s.hosts_fingerprints[in.Fingerprint]
	if !ok {
		return &pb.HostFingerprintResp{Message: in.Host + " not exist!", Sucsess: false}, nil
	}
	s.mutex.Lock()
	delete(s.hosts_fingerprints, in.Fingerprint)
	s.mutex.Unlock()

	return &pb.HostFingerprintResp{Message: in.Host + " deleted", Sucsess: true}, nil
}

//Initiate soket connection from Location
func (s *tunnelServer) InitConnection(stream pb.SocketConnection_InitConnectionServer) error {

	peer, _ := peer.FromContext(stream.Context())
	raw := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Raw
	hex_sert_raw := getFingerprint(raw)

	host, ok := s.hosts_fingerprints[hex_sert_raw]
	if !ok {
		cn := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Subject.CommonName
		lg.Error("Strange clienf sert", zap.String("Fingerprint", hex_sert_raw), zap.String("CommonName", cn))
		return errors.New("strange clienf sert:" + cn)
	}

	s.mutex.Lock()
	s.hosts_sokets[host] = stream
	s.mutex.Unlock()
	defer func() {
		s.mutex.Lock()
		delete(s.hosts_sokets, host)
		s.mutex.Unlock()
	}()

	for {
		//receive json from location
		in, err := stream.Recv()
		if err == io.EOF {
			lg.Error("Stream connection lost", zap.String("Host", host))
			s.request_id[in.Id] <- nil
			return err
		}
		if err != nil {
			lg.Error("stream.Recv", zap.String("Host", host))
			s.request_id[in.Id] <- nil
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

//Start http server to pass request to grpc, next Location
func (s *tunnelServer) startHttpServer() *http.Server {
	port := viper.GetString("HTTP_PORT")
	srv := &http.Server{Addr: ":" + port}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// r.URL.String() or r.URL.RequestURI() show only r.URL.Path!
		// r.URL.Scheme is empty! Possible  r.TLS == nil

		host_soket, ok := s.hosts_sokets[r.Host]
		// host_soket, ok := s.hosts_sokets["httpbin.org"]
		if !ok {
			lg.Error("Connection not found", zap.String("GetHost", r.Host), zap.String("f", "SendData"))
			http.Error(w, " not found", http.StatusInternalServerError)
			return
		}
		//send json to location
		body, err := io.ReadAll(r.Body)
		if err != nil {
			lg.Error("Failed io.ReadAll", zap.String("Host", r.Host))
			http.Error(w, "Failed io.ReadAll", http.StatusInternalServerError)
			return
		}

		req_struct := pb.HttpData{
			Status: 200, Method: r.Method, Path: r.URL.Path, Host: r.Host, Header: r.Header, Body: body}
		//TODO Zip json_req?
		json_req, err := json.Marshal(req_struct)
		if err != nil {
			lg.Error("json.Marshal", zap.String("Host", r.Host))
			http.Error(w, "json.Marshal", http.StatusInternalServerError)
			return
		}

		//request ID
		//uint32	0 — 4 294 967 295
		rand.Seed(time.Now().UnixNano())
		rnd := rand.Uint32()

		s.mutex.Lock()
		s.request_id[rnd] = make(chan []byte)
		s.mutex.Unlock()
		defer func() {
			s.mutex.Lock()
			delete(s.request_id, rnd)
			s.mutex.Unlock()
		}()

		err = host_soket.Send(&pb.HttpReQuest2Loc{Id: rnd, Message: r.Host, Json: json_req})
		if err != nil {
			lg.Error("Failed stream.Send", zap.String("Host", r.Host))
			http.Error(w, "Failed stream.Send", http.StatusInternalServerError)
			return
		}

		lg.Info("Sended data to:", zap.String("Host", r.Host))

		select {
		case inJson := <-s.request_id[rnd]:
			if inJson == nil {
				http.Error(w, "Connection closed", http.StatusServiceUnavailable)
			} else {
				var resp pb.HttpData
				err = json.Unmarshal(inJson, &resp)
				if err != nil {
					lg.Error("json.Unmarshal", zap.String("Host", r.Host))
					http.Error(w, "json.Unmarshal", http.StatusInternalServerError)
					return
				}

				w.WriteHeader(resp.Status)
				w.Write([]byte(resp.Body))
			}
		case <-time.After(30 * time.Second): // for close and defer delete(s.request_id, rnd)
			http.Error(w, "Failed stream.Recv, timeout", http.StatusGatewayTimeout)
		}

	})

	go func() {
		fmt.Println("startHttpServer")
		lg.Info("StartHttpServer on localhost:", zap.String("port", port), zap.Skip())
		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexected error. port in use?
			lg.Fatal("failed to serve grpc:", zap.Error(err))
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func main() {
	port := viper.GetString("GRPC_PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.Fatal("failed to listen:", zap.Error(err))
	}

	var opts []grpc.ServerOption

	if viper.GetBool("SECURE") {
		//openssl req -new -newkey rsa:4096 -x509 -sha256 -days 30 -nodes -out server.crt -keyout server.key
		cert, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
		if err != nil {
			lg.Fatal("server: loadkeys:", zap.Error(err))
		}

		// Load CA cert
		caCertPool := x509.NewCertPool()
		//Certification authority, CA
		//A CA certificate is a digital certificate issued by a certificate authority (CA), so SSL clients (such as web browsers) can use it to verify the SSL certificates sign by this CA.
		//todo собрать клиентские сертификаты в один файл
		caCert, err := ioutil.ReadFile("cert/0client.crt")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
		caCert, err = ioutil.ReadFile("cert/1client.crt")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool.AppendCertsFromPEM(caCert)

		// Note if we don't tls.RequireAnyClientCert client side certs are ignored.
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// ClientCAs:    caCertPool,//It work! Peer client sertificates autenification
			ClientAuth: tls.RequireAnyClientCert,
			//ClientAuth: tls.RequireAndVerifyClientCert,
			RootCAs: caCertPool,
			// VerifyPeerCertificate: customVerify,
			// VerifyPeerCertificate: is called only after normal certificate verification https://pkg.go.dev/crypto/tls#Config
			InsecureSkipVerify: false,
			//InsecureSkipVerify: true,
		}
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.Creds(cred))
		// // Enable TLS for all incoming connections.
		// opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
		// failed to complete security handshake on grpc?
		// https://stackoverflow.com/questions/43829022/failed-to-complete-security-handshake-on-grpc
	}

	// opts = append(opts, grpc.StreamInterceptor(streamInterceptor))
	//
	grpcServer := grpc.NewServer(opts...)

	newServiceImpl := &tunnelServer{
		hosts_fingerprints: make(map[string]string),
		hosts_sokets:       make(map[string]pb.SocketConnection_InitConnectionServer),
		request_id:         make(map[uint32]chan ([]byte)),
	}

	newServiceImpl.getHostFingerprintsFromDB()

	pb.RegisterSocketConnectionServer(grpcServer, newServiceImpl)

	srv := newServiceImpl.startHttpServer()

	// pb.RegisterTunnelServer(grpcServer, &tunnelServer{})
	lg.Info("gRPC-Server Listening on localhost:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		lg.Fatal("failed to serve grpc:", zap.Error(err))
	}

	srv.Shutdown(context.TODO())
}
