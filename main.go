package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
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

	r.HandleFunc("/string", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Access-Control-Allow-Origin", "https://stt.sasu.net")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		var src cryptoSource
		rnd := rand.New(src)

		//vars := mux.Vars(r)
		//string := vars["string"]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		string := string(body)
		fmt.Println("Received request POST /string with body " + string)
		if string == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		//TODO: check that pin does not exist already
		pin := uint64(rnd.Intn(99999))
		m[pin] = string

		sPin := strconv.FormatUint(pin, 10)
		fmt.Println("Returning created PIN " + sPin)
		io.WriteString(w, sPin)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/pin", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		//vars := mux.Vars(r)
		//pinStr := vars["pin"]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		pinStr := string(body)
		fmt.Println("Received request POST /pin with body " + pinStr)

		pin, err := strconv.ParseUint(pinStr, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		pwd := m[pin]
		if pwd == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		delete(m, pin)
		fmt.Println("Returning retrieved string " + pwd)
		io.WriteString(w, pwd)
	}).Methods("POST", "OPTIONS")

	http.ListenAndServe(":10000", r)
}

/*
NotFoundHandler returns nothing
*/
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
