package main

import (
	crand "crypto/rand"
	"encoding/binary"
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
		var src cryptoSource
		rnd := rand.New(src)

		vars := mux.Vars(r)
		string := vars["string"]
		fmt.Println("Received request POST /string/" + string)

		pin := uint64(rnd.Intn(99999))
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

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
