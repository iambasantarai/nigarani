package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./ui")))

	fmt.Printf("Application started.\nLink: http://localhost:%d\n", 8000)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("ERROR: %s", err.Error())
	}
}
