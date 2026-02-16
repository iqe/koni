package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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

type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

func apacheLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
			clientIP = clientIP[:colon]
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		startTime := time.Now()
		next.ServeHTTP(rec, r)
		finishTime := time.Now()

		record := &apacheLogRecord{
			ip:            r.RemoteAddr,
			time:          finishTime.UTC(),
			method:        r.Method,
			uri:           r.RequestURI,
			protocol:      r.Proto,
			status:        rec.status,
			elapsedTime:   finishTime.Sub(startTime),
			responseBytes: int64(rec.size),
		}

		timeFormatted := record.time.Format(apacheLogTimeFormat)
		requestLine := fmt.Sprintf("%s %s %s", record.method, record.uri, record.protocol)
		log.Printf(apacheFormatPattern, record.ip, timeFormatted, requestLine, record.status, record.responseBytes, record.elapsedTime.Seconds())
	})
}
