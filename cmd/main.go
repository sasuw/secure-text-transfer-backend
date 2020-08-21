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
	"time"

	"github.com/gorilla/mux"
)

var m map[uint64]string

func main() {

	m = make(map[uint64]string)

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	r.HandleFunc("/string", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		string := string(body)
		//TODO: make environment-specific logging for debugging purposes
		//fmt.Println("Received request POST /string with body " + string)
		if string == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		pin := GetPin()
		if pin == 0 {
			//no pin could be generated
			fmt.Println("Error: PIN could not be generated")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		m[pin] = string

		time.AfterFunc(time.Minute*5, func() {
			if m[pin] != "" {
				log.Println("Auto-deleted text for PIN " + strconv.FormatUint(pin, 10))
				delete(m, pin)
			}
		})

		sPin := strconv.FormatUint(pin, 10)
		fmt.Println("Returning created PIN " + sPin)
		io.WriteString(w, sPin)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/pin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		pinStr := string(body)
		//TODO: make environment-specific logging for debugging purposes
		//fmt.Println("Received request POST /pin with body " + pinStr)

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

	http.ListenAndServe(":9999", r)
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

/*
GetPin returns a PIN of type uint64 with a value between 1 and 99999
*/
func GetPin() uint64 {
	var src cryptoSource
	rnd := rand.New(src)

	pin := uint64(rnd.Intn(99999))
	var i int = 0
	var numberOfTimesToTry = 100 //100 subsequent collisions should be very unlikely
	valuePresent := false
	keepTrying := true
	for ok := true; ok; ok = keepTrying {
		_, valuePresent = m[pin]
		keepTrying = valuePresent && i < numberOfTimesToTry
		i++
	}

	if !valuePresent {
		return pin
	}

	return 0
}
