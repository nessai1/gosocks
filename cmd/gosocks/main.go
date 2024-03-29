package main

import (
	"flag"
	"gosocks/internal/gosocks"
	"log"
)

func main() {
	addr := flag.String("a", ":1080", "Address of proxy")
	flag.Parse()

	err := gosocks.ListenAndServe(*addr)
	if err != nil {
		log.Fatalf("Error while listen gosocks: %s", err.Error())
	}
}
