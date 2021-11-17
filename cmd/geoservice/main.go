package main

import (
	"fmt"
	"geoservice/internal/server"
	"log"
)

var (
	port string = server.GetPort()
)

func main() {
	fmt.Println("Starting server...")
	fmt.Println("Processing GeoJSON files. This normally takes a bit.")
	srv := server.NewHTTPServer("0.0.0.0:" + port)
	fmt.Println("Processing complete.")
	fmt.Printf("Server listening at port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
