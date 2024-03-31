package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nessai1/gosocks/internal/gosocks"
	"github.com/nessai1/gosocks/internal/storage"
	"log"
)

func main() {
	addr := flag.String("a", ":1080", "Address of proxy")
	hasAuth := flag.Bool("auth", false, "Use auth to proxy")

	// TODO: need to change add algorithm
	newUsername := flag.String("username", "", "Username of new client")
	newPassword := flag.String("password", "", "Password of new client")

	flag.Parse()
	var store storage.Storage = nil
	var err error = nil

	if *newUsername != "" && *newPassword != "" {
		store, err = createStorage()
		if err != nil {
			panic(fmt.Sprintf("Cannot create new client: %s", err.Error()))
		}

		client, err := store.CreateClient(context.TODO(), *newUsername, *newPassword)
		if err != nil {
			panic(fmt.Sprintf("Cannot create new client: %s", err.Error()))
		}

		fmt.Printf("Client %s successful created", client.Login)
		return
	}

	if *hasAuth {
		store, err = createStorage()
		if err != nil {
			log.Fatalf("Cannot create storage: %s", err.Error())
		}
	}

	err = gosocks.ListenAndServe(*addr, store)
	if err != nil {
		log.Fatalf("Error while listen gosocks: %s", err.Error())
	}
}

func createStorage() (storage.Storage, error) {
	s, err := storage.NewSQLiteStorage("storage")
	if err != nil {
		return nil, fmt.Errorf("cannot create storage: %w", err)
	}

	return s, nil
}
