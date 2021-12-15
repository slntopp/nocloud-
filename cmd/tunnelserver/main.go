package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"

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

type tunnelServerParams struct {
	stream pb.SocketConnection_InitConnectionServer
	ch     chan error
}

type tunnelServer struct {
	pb.UnimplementedSocketConnectionServer
	tsps map[string]tunnelServerParams
}

func getFingerprint(c []byte) []byte {
	sum := sha256.Sum256(c)
	return sum[:]
}

func getHostnameFromDB(c string) string {
	if c == "419d3335b2b533526d4e7f6f1041b3c492d086cad0f5876739800ffd51659545" {
		// return "zero.client.net"
		return "ione-cloud.net"
	}
	return ""
}

//Initiate soket connection from Location
func (s *tunnelServer) InitConnection(stream pb.SocketConnection_InitConnectionServer) error {

	peer, _ := peer.FromContext(stream.Context())
	raw := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Raw
	hex_sert_raw := hex.EncodeToString(getFingerprint(raw))

	host := getHostnameFromDB(hex_sert_raw)
	s.tsps[host] = tunnelServerParams{stream, make(chan error)}
	defer delete(s.tsps, host)

	return <-s.tsps[host].ch
}

//Start http server to pass request to grpc, next Location
func (s *tunnelServer) startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":" + viper.GetString("HTTP_PORT")}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		tsp, ok := s.tsps[r.Host]
		if !ok {
			lg.Error("Connection not found", zap.String("GetHost", r.Host), zap.String("f", "SendData"))
			http.Error(w, "GRCP connection not found", http.StatusInternalServerError)
			return
		}
		//send json to location
		body, _ := io.ReadAll(r.Body)

		// r.URL.String() or r.URL.RequestURI()
		//show only r.URL.Path!
		var FullUrl string
		if r.TLS == nil {
			FullUrl = "http://"
		} else {
			FullUrl = "https://"
		}
		FullUrl += r.Host + r.URL.Path

		req_struct := pb.HttpData{
			Status: 200, Method: r.Method, FullURL: FullUrl, Host: r.Host, Header: r.Header, Body: body}
		//TODO Zip json_req?
		json_req, err := json.Marshal(req_struct)
		if err != nil {
			lg.Error("json.Marshal", zap.String("Host", r.Host))
			http.Error(w, "json.Marshal", http.StatusInternalServerError)
			return
		}

		err = tsp.stream.Send(&pb.HttpReQuest2Loc{Message: r.Host, Json: json_req})
		if err != nil {
			lg.Error("Failed stream.Send", zap.String("Host", r.Host))
			tsp.ch <- err
			http.Error(w, "Failed stream.Send", http.StatusInternalServerError)
			return
		}
		//receive json from location
		in, err := tsp.stream.Recv()
		if err == io.EOF {
			lg.Error("Stream connection lost", zap.String("Host", r.Host))
			tsp.ch <- nil
			http.Error(w, "Stream connection lost", http.StatusInternalServerError)
			return
		}
		if err != nil {
			lg.Error("stream.Recv", zap.String("Host", r.Host))
			tsp.ch <- err
			http.Error(w, "stream.Recv", http.StatusInternalServerError)
			return
		}

		var resp pb.HttpData
		err = json.Unmarshal(in.Json, &resp)
		if err != nil {
			lg.Error("json.Unmarshal", zap.String("Host", r.Host))
			tsp.ch <- err
			http.Error(w, "json.Unmarshal", http.StatusInternalServerError)
			return
		}

		lg.Info("IS:", zap.String("Host", r.Host))

		w.WriteHeader(resp.Status)
		w.Write([]byte(resp.Body))

	})

	go func() {
		fmt.Println("startHttpServer")
		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
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

	grpcServer := grpc.NewServer(opts...)

	newServiceImpl := &tunnelServer{tsps: make(map[string]tunnelServerParams)}
	pb.RegisterSocketConnectionServer(grpcServer, newServiceImpl)

	srv := newServiceImpl.startHttpServer()

	// pb.RegisterTunnelServer(grpcServer, &tunnelServer{})
	lg.Info("gRPC-Server Listening on localhost:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		lg.Fatal("failed to serve grpc:", zap.Error(err))
	}

	srv.Shutdown(context.TODO())
}
