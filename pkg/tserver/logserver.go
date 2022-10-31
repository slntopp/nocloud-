package tserver

import (
	"errors"
	"io"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"go.uber.org/zap"
)

var (
	startLogConnectionStriam chan (string)
)

func init() {
	startLogConnectionStriam = make(chan (string))
}

// Getting container logs
func (s *TunnelServer) LogConnection(stream pb.SocketConnectionService_LogConnectionServer) error {
	log := s.log.Named("LogConnection")

	host := <-startLogConnectionStriam
	if host == "" {
		return errors.New("Unregistered Certificate")
	}

	log.Info("Client connected", zap.String("Host", host))

	s.mutex.Lock()
	newTH := s.hosts[host]
	newTH.streamLog = stream
	s.hosts[host] = newTH
	s.mutex.Unlock()

	go func() {
		var err_global error
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
				err_global = err
				break
			}

			err = s.streamLogAdmin.Send(&pb.LogResponse{Log: in.Log})
			if err != nil {
				err_global = err
				break
			}

			log.Debug("Received data from:", zap.String("Host", host))
		}

		s.hosts[host].resetConn <- err_global
	}()

	err_global := <-s.hosts[host].resetConn
	log.Info("Connection closed", zap.String("Host", host), zap.Error(err_global))
	return err_global
}

// Getting container logs
func (s *TunnelServer) Log(stream pb.SocketConnectionService_LogAdminServer) error {
	log := s.log.Named("Log")

	//TODO check autenification
	s.streamLogAdmin = stream

	log.Info("Log reader connected")

	var err_global error
	for {
		//receive json from location
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Error("Log reader Stream connection lost")
			} else {
				log.Error("stream.Recv", zap.Error(err))
			}
			err_global = err
			break
		}

		log.Debug("Log reader Receive data to:", zap.String("Host", in.Host))

		if th, ok := s.hosts[in.Host]; ok {
			th.streamLog.Send(&pb.LogConnectionResponse{
				TsStart: in.TsStart,
				Follow:  in.Follow,
				Stop:    in.Stop})
		} else {
			err := stream.Send(&pb.LogResponse{Err: "Host disabled"})
			if err != nil {
				err_global = err
				break
			}
		}

	}

	log.Info("Log reader connection closed", zap.Error(err_global))
	return err_global
}
