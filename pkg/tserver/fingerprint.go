package tserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
)

//Make hash od raw sertificate as sha256
func MakeFingerprint(c []byte) string {
	sum := sha256.Sum256(c)
	return hex.EncodeToString(sum[:])
}
//Gets fingerprint-host chain from DB
func (s *TunnelServer) LoadHostFingerprintsFromDB() {
	fp := "419d3335b2b533526d4e7f6f1041b3c492d086cad0f5876739800ffd51659545"
	s.fingerprints_hosts[fp] = "demo.ione.local"
	// s.hosts_fingerprints[fp] ="reqbin.com"
	// s.hosts_fingerprints[fp] ="zero.client.net"
	// s.hosts_fingerprints[fp] ="ione-cloud.net"
}

//Add host/Fingerprint to DB
func (s *TunnelServer) Add(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {

	_, ok := s.fingerprints_hosts[in.Fingerprint]
	if ok {
		return &pb.HostFingerprintResp{Message: in.Host + " exists!", Sucsess: false}, nil
	}

	s.mutex.Lock()
	s.fingerprints_hosts[in.Fingerprint] = in.Host
	s.mutex.Unlock()

	return &pb.HostFingerprintResp{Message: in.Host + " added", Sucsess: true}, nil
}

//Edit host/Fingerprint to DB
func (s *TunnelServer) Edit(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {

	_, ok := s.fingerprints_hosts[in.Fingerprint]
	if !ok {
		return &pb.HostFingerprintResp{Message: in.Host + " not exist!", Sucsess: false}, nil
	}
	s.mutex.Lock()
	s.fingerprints_hosts[in.Fingerprint] = in.Host
	s.mutex.Unlock()

	return &pb.HostFingerprintResp{Message: in.Host + " edited", Sucsess: true}, nil
}

//Delete host/Fingerprint to DB
func (s *TunnelServer) Delete(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {
	s.fingerprints_hosts[in.Fingerprint] = in.Host

	_, ok := s.fingerprints_hosts[in.Fingerprint]
	if !ok {
		return &pb.HostFingerprintResp{Message: in.Host + " not exist!", Sucsess: false}, nil
	}
	s.mutex.Lock()
	delete(s.fingerprints_hosts, in.Fingerprint)
	s.mutex.Unlock()

	return &pb.HostFingerprintResp{Message: in.Host + " deleted", Sucsess: true}, nil
}
