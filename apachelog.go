package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/macaron.v1"
)

const (
	apacheFormatPattern = "%s - - [%s] \"%s %d %d\" %f\n"
	apacheLogTimeFormat = "02/Jan/2006 03:04:05"
)

type apacheLogRecord struct {
	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
}

func apacheLogHandler(ctx *macaron.Context, log *log.Logger) {
	clientIP := ctx.Req.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	rw := ctx.Resp.(macaron.ResponseWriter)

	startTime := time.Now()
	ctx.Next()
	finishTime := time.Now()

	r := &apacheLogRecord{
		ip:            ctx.Req.RemoteAddr,
		time:          finishTime.UTC(),
		method:        ctx.Req.Method,
		uri:           ctx.Req.RequestURI,
		protocol:      ctx.Req.Proto,
		status:        rw.Status(),
		elapsedTime:   finishTime.Sub(startTime),
		responseBytes: int64(rw.Size()),
	}

	timeFormatted := r.time.Format(apacheLogTimeFormat)
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	log.Printf(apacheFormatPattern, r.ip, timeFormatted, requestLine, r.status, r.responseBytes, r.elapsedTime.Seconds())
}
