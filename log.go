package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/macaron.v1"
)

const (
	ApacheFormatPattern = "%s - - [%s] \"%s %d %d\" %f\n"
	ApacheLogTimeFormat = "02/Jan/2006 03:04:05"
)

type ApacheLogRecord struct {
	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
}

func (r *ApacheLogRecord) Log(out io.Writer) {
	timeFormatted := r.time.Format("02/Jan/2006 03:04:05")
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	fmt.Fprintf(out, ApacheFormatPattern, r.ip, timeFormatted, requestLine, r.status, r.responseBytes,
		r.elapsedTime.Seconds())
}

func apacheLogHandler(ctx *macaron.Context) {
	clientIP := ctx.Req.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	rw := ctx.Resp.(macaron.ResponseWriter)

	startTime := time.Now()
	ctx.Next()
	finishTime := time.Now()

	record := &ApacheLogRecord{
		ip:          ctx.Req.RemoteAddr,
		time:        finishTime.UTC(),
		method:      ctx.Req.Method,
		uri:         ctx.Req.RequestURI,
		protocol:    ctx.Req.Proto,
		status:      rw.Status(),
		elapsedTime: finishTime.Sub(startTime),
	}

	record.Log(os.Stdout)
}

// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func Logger() macaron.Handler {
	return func(ctx *macaron.Context, log *log.Logger) {
		start := time.Now()

		log.Printf("%s: Started %s %s for %s", time.Now().Format(macaron.LogTimeFormat), ctx.Req.Method, ctx.Req.RequestURI, ctx.RemoteAddr())

		rw := ctx.Resp.(macaron.ResponseWriter)
		ctx.Next()

		content := fmt.Sprintf("%s: Completed %s %s %v %s in %v", time.Now().Format(macaron.LogTimeFormat), ctx.Req.Method, ctx.Req.RequestURI, rw.Status(), http.StatusText(rw.Status()), time.Since(start))
		if macaron.ColorLog {
			switch rw.Status() {
			case 200, 201, 202:
				content = fmt.Sprintf("\033[1;32m%s\033[0m", content)
			case 301, 302:
				content = fmt.Sprintf("\033[1;37m%s\033[0m", content)
			case 304:
				content = fmt.Sprintf("\033[1;33m%s\033[0m", content)
			case 401, 403:
				content = fmt.Sprintf("\033[4;31m%s\033[0m", content)
			case 404:
				content = fmt.Sprintf("\033[1;31m%s\033[0m", content)
			case 500:
				content = fmt.Sprintf("\033[1;36m%s\033[0m", content)
			}
		}
		log.Println(content)
	}
}
