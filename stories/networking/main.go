package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello! the time right now is %v\n", time.Now())
}

func main() {
	http.HandleFunc("/", handler)

	log.Println("booting up server...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
