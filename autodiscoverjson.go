package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type autodiscoverjsonRedirect struct {
	Protocol string
	Url      string
}

type autodiscoverjsonError struct {
	ErrorCode    string
	ErrorMessage string
}

var autodiscoverV1Protocol = regexp.MustCompile(`(?i)^AutodiscoverV1$`)

func autodiscoverjsonHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		protocol := r.URL.Query().Get("Protocol")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if autodiscoverV1Protocol.MatchString(protocol) {
			url := fmt.Sprintf("https://%s/Autodiscover/Autodiscover.xml", r.Host)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&autodiscoverjsonRedirect{Protocol: "AutodiscoverV1", Url: url})
		} else {
			message := fmt.Sprintf("The given protocol value '%s' is invalid. Supported values are: AutodiscoverV1", protocol)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&autodiscoverjsonError{ErrorCode: "InvalidProtocol", ErrorMessage: message})
		}
	}
}
