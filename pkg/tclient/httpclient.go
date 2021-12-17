package tclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"go.uber.org/zap"
)

func HttpClient(logger *zap.Logger, dest string, stream pb.SocketConnection_InitConnectionClient, message string, id uint32, inJson []byte) {
	log := logger.Named("HTTPClient")
	
	//-------http client
	var req_struct pb.HttpData
	err := json.Unmarshal(inJson, &req_struct)
	if err != nil {
		log.Error("json.Unmarshal", zap.String("Message", message))
		return
	}

	url := dest + req_struct.Path
	req_client, err := http.NewRequest(
		req_struct.Method,
		url,
		bytes.NewBuffer([]byte(req_struct.Body)))
	if err != nil {
		log.Error("http.NewRequest", zap.String("Message", message))
		return
	}
	
	req_client.Header = req_struct.Header

	var resp_struct pb.HttpData

	var netClient = &http.Client{
		Timeout: time.Second * 20,
		// Transport: &http.Transport{
		// 	TLSClientConfig: &tls.Config{
		// 		InsecureSkipVerify: true,
		// 	},
		// },
	}

	response, err := netClient.Do(req_client)
	if err != nil {
		log.Error("Failed to get http:", zap.Error(err))
		//Return error and response==nil if server cant request.
		//No error if server sended any status

		resp_struct = pb.HttpData{
			Status: http.StatusBadRequest,
			Method: req_struct.Method,
			Header: req_struct.Header,
			Body:   []byte(err.Error()),
		}

	} else {

		bodyS, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error("Failed to read responce", zap.Error(err))
			return
		}

		stt, err := strconv.Atoi(response.Status[0:3])
		if err != nil {
			log.Error("strconv.Atoi", zap.Error(err))
		}

		resp_struct = pb.HttpData{
			Status: stt,
			Method: response.Request.Method,
			Header: response.Header,
			Body:   bodyS,
		}

	}
	//-------http client

	jsonResp, err := json.Marshal(resp_struct)
	if err != nil {
		log.Error("json.Marshal", zap.Error(err))
		return
	}

	if err := stream.Send(&pb.HttpReSp4Loc{
		Id:      id,
		Message: req_struct.Host,
		Json:    jsonResp}); err != nil {
		log.Error("Failed to send a note:", zap.Error(err))
		return
	}

}
