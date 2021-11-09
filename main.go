package main

import (
	"fmt"
	"geodb/server"
	"log"
)

const (
	port string = "8083"
)

func main() {
	fmt.Println("Starting server...")
	fmt.Println("Processing GeoJSON files. This normally takes a bit.")
	srv := server.NewHTTPServer(":" + port)
	fmt.Println("Processing complete.")
	fmt.Printf("Server listening at port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
