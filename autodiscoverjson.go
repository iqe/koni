package main

import (
	"fmt"
	"regexp"

	macaron "gopkg.in/macaron.v1"
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

func autodiscoverjsonHandler() macaron.Handler {
	return func(ctx *macaron.Context) {
		protocol := ctx.Req.URL.Query().Get("Protocol")
		if autodiscoverV1Protocol.MatchString(protocol) {
			url := fmt.Sprintf("https://%s/Autodiscover/Autodiscover.xml", ctx.Req.Host)
			ctx.JSON(200, &autodiscoverjsonRedirect{Protocol: "AutodiscoverV1", Url: url})
		} else {
			message := fmt.Sprintf("The given protocol value '%s' is invalid. Supported values are: AutodiscoverV1", protocol)
			ctx.JSON(400, &autodiscoverjsonError{ErrorCode: "InvalidProtocol", ErrorMessage: message})
		}
	}
}
