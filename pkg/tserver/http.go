package tserver

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"time"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"go.uber.org/zap"
)

//Start http server to pass request to grpc, then next Location
func (s *TunnelServer) StartHttpServer() *http.Server {
	log := s.log.Named("HTTPServer")
	srv := &http.Server{Addr: ":80"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// r.URL.String() or r.URL.RequestURI() show only r.URL.Path!
		// r.URL.Scheme is empty! Possible  r.TLS == nil

		host_soket, ok := s.hosts[r.Host]
		if !ok {
			log.Error("Connection not found", zap.String("GetHost", r.Host), zap.String("f", "SendData"))
			http.Error(w, " not found", http.StatusInternalServerError)
			return
		}
		//send json to location
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("Failed io.ReadAll", zap.String("Host", r.Host))
			http.Error(w, "Failed io.ReadAll", http.StatusInternalServerError)
			return
		}

		req_struct := pb.HttpData{
			Status: 200, Method: r.Method, Path: r.URL.Path, Host: r.Host, Header: r.Header, Body: body}
		//TODO Zip json_req?
		json_req, err := json.Marshal(req_struct)
		if err != nil {
			log.Error("json.Marshal", zap.String("Host", r.Host))
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

		err = host_soket.stream.Send(&pb.HttpReQuest2Loc{Id: rnd, Message: r.Host, Json: json_req})
		if err != nil {
			log.Error("Failed stream.Send", zap.String("Host", r.Host))
			http.Error(w, "Failed stream.Send", http.StatusInternalServerError)
			return
		}

		log.Info("Sended data to:", zap.String("Host", r.Host))

		select {
		case inJson := <-s.request_id[rnd]:
			log.Info("Get request from client", zap.String("Host", r.Host))
			var resp pb.HttpData
			err = json.Unmarshal(inJson, &resp)
			if err != nil {
				log.Error("json.Unmarshal", zap.String("Host", r.Host))
				http.Error(w, "json.Unmarshal", http.StatusInternalServerError)
				return
			}
			for key, values := range resp.Header {
				for _, val := range values {
					w.Header().Add(key, val)
				}
			}
			w.WriteHeader(resp.Status) //if you have w.WriteHeader(200) before w.Header the headers will not be set!!!
			w.Write(resp.Body)
			log.Info("Send request to http", zap.String("Host", r.Host))

		case <-host_soket.ctx.Done():
			log.Info("Stream.Recv closed", zap.String("Host", r.Host))
			http.Error(w, "Stream.Recv closed", http.StatusInternalServerError)
		case <-r.Context().Done():
			log.Info("HTTP request cancelled", zap.String("Host", r.Host))
			http.Error(w, "HTTP request cancelled", http.StatusGatewayTimeout)
		case <-time.After(30 * time.Second): // for close and defer delete(s.request_id, rnd)
			log.Info("Failed stream.Recv, timeout", zap.String("Host", r.Host))
			http.Error(w, "Failed stream.Recv, timeout", http.StatusGatewayTimeout)
		}
	})

	go func() {
		log.Info("StartHttpServer on 0.0.0.0:80")
		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexected error. port in use?
			log.Fatal("failed to serve grpc:", zap.Error(err))
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}
