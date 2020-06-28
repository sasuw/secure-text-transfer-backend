package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var m map[uint64]string

func main() {

	m = make(map[uint64]string)

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	r.HandleFunc("/string/{string}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		string := vars["string"]
		fmt.Println("Received request POST /string/" + string)

		pin := uint64(rand.Intn(10000))
		m[pin] = string

		sPin := strconv.FormatUint(pin, 10)
		fmt.Println("Returning created PIN " + sPin)
		io.WriteString(w, sPin)
	}).Methods("POST")

	r.HandleFunc("/pin/{pin}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pinStr := vars["pin"]
		fmt.Println("Received request GET /pin/" + pinStr)

		pin, err := strconv.ParseUint(pinStr, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		pwd := m[pin]
		if pwd == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		fmt.Println("Returning retrieved string " + pwd)
		io.WriteString(w, pwd)
	}).Methods("GET")

	http.ListenAndServe(":10000", r)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "404 - Not Found")
}
