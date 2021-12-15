package proto

import (
	"net/http"
)

//sruct for pass data by json
type HttpData struct {
	Status  int
	Method  string
	Path string
	Host string
	Header  http.Header
	Body    []byte
}
