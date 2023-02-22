package main

import (
	"log"
	"net/http"

	"github.com/wagnerlduarte/hellogolang/routers"
)

// gin --appPort 8080 --port 3000
func main() {
	mux := routers.ConfigureEndpoints()

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatal(err)
	}
}
