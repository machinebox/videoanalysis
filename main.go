package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/machinebox/sdk-go/facebox"
	"github.com/machinebox/sdk-go/x/boxutil"
)

var (
	addr = flag.String("addr", ":8081", "address")
)

func main() {
	flag.Parse()
	facebox := facebox.New("http://localhost:8080")
	fmt.Println(`Video Analytics by Machine Box - https://machinebox.io/`)

	fmt.Println("Waiting for Facebox to be ready...")
	boxutil.WaitForReady(context.Background(), facebox)
	fmt.Println("Done!")

	fmt.Println("Setup facebox state")

	fmt.Println("Go to:", *addr+"...")

	srv := NewServer("./assets", "./videos", facebox)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		log.Fatalln(err)
	}
}
