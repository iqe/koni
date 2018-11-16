package main

import (
	"log"
	"net/http/httputil"

	macaron "gopkg.in/macaron.v1"
)

func debugLogHandler() macaron.Handler {
	return func(ctx *macaron.Context) {
		dump, err := httputil.DumpRequest(ctx.Req.Request, true)
		if err != nil {
			log.Printf("koni: Failed to dump request: %v\n", err)
			return
		}

		log.Println("")
		log.Printf("%s", dump)
		log.Println("")
	}
}
