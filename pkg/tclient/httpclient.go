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

func httpClient(stream pb.SocketConnection_InitConnectionClient, message string, id uint32, inJson []byte) {
	//-------http client
	var req_struct pb.HttpData
	err := json.Unmarshal(inJson, &req_struct)
	if err != nil {
		lg.Error("json.Unmarshal", zap.String("Message", message))
		return
	}

	url := DESTINATION_HOST + req_struct.Path
	req_client, err := http.NewRequest(
		req_struct.Method,
		url,
		bytes.NewBuffer([]byte(req_struct.Body)))
	if err != nil {
		lg.Error("http.NewRequest", zap.String("Message", message))
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
		lg.Error("Failed to get http:", zap.Error(err))
		//Return error and response==nil if server cant request.
		//No error if server sended any status

		resp_struct = pb.HttpData{
			Status: http.StatusBadRequest,
			Method: req_struct.Method,
			// FullURL:    *response.Request.URL,
			Header: req_struct.Header,
			Body:   []byte(err.Error()),
		}

	} else {

		bodyS, err := ioutil.ReadAll(response.Body)
		if err != nil {
			lg.Error("Failed to read responce", zap.Error(err))
			return
		}

		stt, err := strconv.Atoi(response.Status[0:3])
		if err != nil {
			lg.Error("strconv.Atoi", zap.Error(err))
		}

		resp_struct = pb.HttpData{
			Status: stt,
			Method: response.Request.Method,
			// FullURL:    *response.Request.URL,
			Header: response.Header,
			Body:   bodyS,
		}

	}
	//-------http client

	jsonResp, err := json.Marshal(resp_struct)
	if err != nil {
		lg.Error("json.Marshal", zap.Error(err))
		return
	}

	if err := stream.Send(&pb.HttpReSp4Loc{
		Id:      id,
		Message: req_struct.Host,
		Json:    jsonResp}); err != nil {
		lg.Error("Failed to send a note:", zap.Error(err))
		return
	}

}
