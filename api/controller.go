package api

import (
	"IOCService/service/checker"
	"encoding/json"
	"log"
	"net/http"
)

type Attributes struct {
	Attributes []string `json:"attributes"`
}

//action: "CheckAttributes"
func CheckAttributes(responseWriter http.ResponseWriter, request *http.Request) {
	var attr Attributes
	if err := readBody(request, &attr); err != nil {
		http.Error(responseWriter, "Data is not format", http.StatusConflict)
		return
	}

	iocs, err := checker.Check(attr.Attributes)
	if err != nil {
		log.Printf("Unexpected error")
	}

	data := map[string][]int64{
		"ioc_ids": iocs,
	}

	response, err := json.Marshal(data)
	if err != nil {
		log.Printf("Unexpected error")
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	_, err = responseWriter.Write(response)
	if err != nil {
		log.Printf("Unexpected error")
	}
}

func readBody(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(&v)
}
