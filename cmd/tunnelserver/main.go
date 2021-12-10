package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

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
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("SECURE", true)
}

type tunnelServerParams struct {
	stream pb.Tunnel_InitTunnelServer
	ch     chan error
}

type tunnelServer struct {
	pb.UnimplementedTunnelServer

	tsps map[string]tunnelServerParams
}

func (s *tunnelServer) SendData(ctx context.Context, req *pb.SendDataRequest) (*pb.SendDataResponse, error) {

	// peer, ok := peer.FromContext(ctx)
	// if ok {

	// 	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	// 	// fmt.Println(tlsInfo)
	// 	// v := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	// 	// fmt.Printf("%v - %v\n", peer.Addr.String(), v)
	// 	//r.TLS.PeerCertificates[0].Subject.CommonName
	// 	for _, v := range tlsInfo.State.PeerCertificates {
	// 		//    fmt.Println("Client: Server public key is:")
	// 		//    fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))

	// 		fmt.Println("Public CommonName is:", v.Subject.CommonName) /*  */
	// 		   fmt.Println("Public Signature is:",v.Signature) //fingerprint?
	// 	}
	// }
	// return &pb.SendDataResponse{Result: "true\n" == req.GetMessage()}, nil

	// fmt.Println(s.tsps)
	tsp, ok := s.tsps[req.GetHost()]
	if !ok {
		lg.Error("Connection not found", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		return nil, status.Error(codes.NotFound, "Connection not found")
	}
	err := tsp.stream.Send(&pb.StreamData{Message: req.GetMessage()})
	if err != nil {
		lg.Error("Failed to send a note:", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		tsp.ch <- err
		return nil, status.Error(codes.Internal, err.Error())
	}

	in, _ := tsp.stream.Recv()
	if err == io.EOF {
		lg.Error("Connection lost", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		tsp.ch <- nil
		return nil, status.Error(codes.NotFound, "Connection lost")
	}
	if err != nil {
		lg.Error("stream.Recv()", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		tsp.ch <- err
		return nil, err
	}

	return &pb.SendDataResponse{Result: "true\n" == in.Host}, nil

}

func (s *tunnelServer) InitTunnel(stream pb.Tunnel_InitTunnelServer) error {

	// peer, ok := peer.FromContext(stream.Context())
	// if ok {
	// 	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	// 	fmt.Println("tlsInfo", tlsInfo.State)
	// 	// fmt.Println("tlsInfo", tlsInfo.State)
	// 	// v := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	// 	// fmt.Printf("%v - %v\n", peer.Addr.String(), v)
	// 	//r.TLS.PeerCertificates[0].Subject.CommonName
	// 	for _, v := range tlsInfo.State.PeerCertificates {
	// 		//    fmt.Println("Client: Server public key is:")
	// 		//    fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))

	// 		v1 := v.Subject.CommonName
	// 		fmt.Println("Public CommonName InitTunnel is:", v1)
	// 	}
	// } else {
	// 	fmt.Println("peer.FromContext: ", ok)
	// }

	// in, err := stream.Recv()
	// if err == io.EOF {
	// 	lg.Error("lost conn InitTunnel:", zap.Error(err), zap.String("f", "InitTunnel"))
	// 	return nil
	// }
	// if err != nil {
	// 	lg.Error("InitTunnel:", zap.Error(err), zap.String("f", "InitTunnel"))
	// 	return err
	// }

	// lg.Info("Hello from client to server!", zap.String("note", in.Host), zap.Skip())

	peer, _ := peer.FromContext(stream.Context())
	cn := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Subject.CommonName
	fmt.Println("Public CommonName InitTunnel is:", cn)

	raw := peer.AuthInfo.(credentials.TLSInfo).State.PeerCertificates[0].Raw
	hex1 := hex.EncodeToString(getFingerprint(raw))
	ggg := sha256.Sum256(raw)
	hex1 = hex.EncodeToString(ggg[:])
	fmt.Println("getFingerprint InitTunnel is:", hex1)

	s.tsps[getHostname(hex1)] = tunnelServerParams{stream, make(chan error)}
	defer delete(s.tsps, cn)

	return <-s.tsps[cn].ch
}

func getFingerprint(c []byte) []byte {
	s := sha256.New()
	_, _ = s.Write(c) // nolint: gosec
	return s.Sum(nil)
}

func getHostname(c string) string {

	if "419d3335b2b533526d4e7f6f1041b3c492d086cad0f5876739800ffd51659545" == c {
		return "zero.client.net"
	}
	return "zero.client.net"
}

// // wrappedStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// // SendMsg method call.
// type wrappedStream struct {
// 	grpc.ServerStream
// }

// func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {

// 	// fmt.Println("s:", s)
// 	return &wrappedStream{s}
// }

// func streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

// 	peer, ok := peer.FromContext(ss.Context())
// 	if ok {
// 		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
// 		fmt.Println("tlsInfo", tlsInfo.State)
// 		fmt.Println("info", info)
// 		// fmt.Println("tlsInfo", tlsInfo.State)
// 		// v := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
// 		// fmt.Printf("%v - %v\n", peer.Addr.String(), v)
// 		//r.TLS.PeerCertificates[0].Subject.CommonName
// 		for _, v := range tlsInfo.State.PeerCertificates {
// 			//    fmt.Println("Client: Server public key is:")
// 			//    fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))

// 			v1 := v.Subject.CommonName
// 			fmt.Println("Public CommonName Interceptor is:", v1)
// 		}
// 	} else {
// 		fmt.Println("peer.FromContext: ", ok)
// 	}

// 	// // authentication (token verification)
// 	// 	md, _ := metadata.FromIncomingContext(ss.Context())
// 	// // 	md: map[:authority:[localhost:8080] content-type:[application/grpc] user-agent:[grpc-go/1.42.0]]
// 	// // info: &{/tunnel.Tunnel/InitTunnel true true}
// 	// fmt.Println("md:", md)
// 	// fmt.Println("info:", info)
// 	// // if !ok {
// 	// // 	return errMissingMetadata
// 	// // }
// 	// // if !valid(md["authorization"]) {
// 	// // 	return errInvalidToken
// 	// // }

// 	handler(srv, newWrappedStream(ss))
// 	// err := handler(srv, newWrappedStream(ss))
// 	// if err != nil {
// 	// 	logger("RPC failed with error %v", err)
// 	// }
// 	return nil
// }

func (s *tunnelServer) startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":8090"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Response params:", r.Host, r.URL.Path)
		// io.WriteString(w, "hello world\n")

		ps := r.URL.Path
		note := ps[strings.LastIndex(ps, "/")+1:]
		tsp, ok := s.tsps[r.Host]
		if !ok {
			lg.Error("Connection not found", zap.String("GetHost", r.Host), zap.String("f", "SendData"))
			http.Error(w, "GRCP connection not found", http.StatusInternalServerError)
			return
		}
		err := tsp.stream.Send(&pb.StreamData{Message: note})
		if err != nil {
			lg.Error("Failed to send a note:", zap.String("Message", note), zap.String("f", "SendData"))
			tsp.ch <- err
			// w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, "Failed to send a note", http.StatusInternalServerError)
			return
		}

		in, err := tsp.stream.Recv()
		if err == io.EOF {
			lg.Error("Connection lost", zap.String("Message", note), zap.String("f", "SendData"))
			tsp.ch <- nil
			http.Error(w, "Connection lost", http.StatusInternalServerError)
			return
		}
		if err != nil {
			lg.Error("stream.Recv()", zap.String("Message", note), zap.String("f", "SendData"))
			tsp.ch <- err
			http.Error(w, "stream.Recv()", http.StatusInternalServerError)
			return
		}

		lg.Info("IS:", zap.String("Message", in.Host), zap.String("f", "SendData"))
		w.Write([]byte(in.Host))

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
	port := viper.GetString("PORT")
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
	pb.RegisterTunnelServer(grpcServer, newServiceImpl)

	srv := newServiceImpl.startHttpServer()

	// pb.RegisterTunnelServer(grpcServer, &tunnelServer{})
	lg.Info("gRPC-Server Listening on localhost:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		lg.Fatal("failed to serve grpc:", zap.Error(err))
	}

	srv.Shutdown(context.TODO())
}
