package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/slntopp/nocloud-tunnel-mesh/pkg/logger"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	lg     *zap.Logger
	host   string
	secure bool
)

func init() {
	lg = logger.NewLogger()

	viper.AutomaticEnv()
	viper.SetDefault("TUNNEL_HOST", "localhost:8080")
	viper.SetDefault("SECURE", true)

	host = viper.GetString("TUNNEL_HOST")
	secure = viper.GetBool("SECURE")
}

func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	if secure {
		// Load client cert
		cert, err := tls.LoadX509KeyPair("cert/0client.crt", "cert/0client.key")
		// cert, err := tls.LoadX509KeyPair("cert/1client.crt", "cert/1client.key")
		if err != nil {
			lg.Fatal("fail to LoadX509KeyPair:", zap.Error(err))
		}

		// Setup HTTPS client
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// InsecureSkipVerify: false,
			InsecureSkipVerify: true,
		}
		// config.BuildNameToCertificate()
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts[0] = grpc.WithTransportCredentials(cred)
	}

	opts = append(opts, grpc.WithBlock())

	//Reconnection
	for {

		func() {
			defer time.Sleep(5 * time.Second)

			lg.Info("Try to connect...", zap.String("host", host), zap.Skip())

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			conn, err := grpc.DialContext(ctx, host, opts...)
			// conn, err := grpc.Dial(host, opts...)
			if err != nil {
				lg.Error("fail to dial:", zap.Error(err))
				return
			}
			defer conn.Close()

			lg.Info("Connected:", zap.String("host", host), zap.Skip())

			client := pb.NewSocketConnectionClient(conn)

			stream, err := client.InitConnection(context.Background()) //, &pb.InitTunnelRequest{Host: "testo"})
			if err != nil {
				lg.Error("Failed InitTunnel", zap.Error(err))
				return
			}

			for {
				in, err := stream.Recv()
				if err == io.EOF {
					lg.Info("Connection closed", zap.String("TUNNEL_HOST", host), zap.Skip())
					return
				}
				if err != nil {
					lg.Error("Failed to receive a note:", zap.Error(err))
					return
				}

				lg.Info("Received httpReQuest2Loc:", zap.String("TUNNEL_HOST", host), zap.Skip())

				//http client
				var req_struct pb.HttpData
				err = json.Unmarshal(in.Json, &req_struct)
				if err != nil {
					lg.Error("json.Unmarshal", zap.String("Host", in.Message))
					return
				}

				req_client, err := http.NewRequest(
					req_struct.Method,
					req_struct.FullURL,
					bytes.NewBuffer([]byte(req_struct.Body)))
				if err != nil {
					lg.Error("http.NewRequest", zap.String("Host", in.Message))
					return
				}

				var netClient = &http.Client{
					Timeout: time.Second * 10,
				}
				//TODO critical/notcritical errors
				response, err := netClient.Do(req_client)
				if err != nil {
					lg.Error("Failed to get http:", zap.Error(err))
					return
				}

				bodyS, err := ioutil.ReadAll(response.Body)
				if err != nil {
					lg.Error("Failed to read responce", zap.Error(err))
					return
				}
				//-------http client

				i, err := strconv.Atoi(response.Status[0:3])
				if err != nil {
					lg.Error("strconv.Atoi")
				}

				resp_struct := pb.HttpData{
					Status: i, //todo
					Method: response.Request.Method,
					// FullURL:    *response.Request.URL,
					Header: response.Header,
					Body:   bodyS,
				}

				jsonResp, _ := json.Marshal(resp_struct)
				if err != nil {
					lg.Error("json.Marshal")
					return
				}

				if err := stream.Send(&pb.HttpReSp4Loc{
					Message: req_struct.Host,
					Json:    jsonResp}); err != nil {
					lg.Error("Failed to send a note:", zap.Error(err))
					return
				}
			}
		}()

	}
}
