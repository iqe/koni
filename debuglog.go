package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func debugLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("koni: Failed to dump request: %v\n", err)
		} else {
			log.Println("")
			log.Printf("%s", dump)
			log.Println("")
		}

		next.ServeHTTP(w, r)
	})
}
