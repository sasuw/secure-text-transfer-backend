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
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var m map[string]string

func main() {

	m = make(map[string]string)

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	r.HandleFunc("/string", func(w http.ResponseWriter, r *http.Request) {
		initResult := InitRequest(w, r)
		if !initResult {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError) //http 500
			return
		}

		dt := time.Now()
		blen := strconv.Itoa(len(body))
		fmt.Println("Received request to store text of length "+blen+" at ", dt.String())

		string := string(body)
		//TODO: make environment-specific logging for debugging purposes
		//fmt.Println("Received request POST /string with body " + string)
		if string == "" {
			w.WriteHeader(http.StatusNoContent) //http 204
			return
		}

		maxTextLength := 4000
		if len(string) > maxTextLength {
			//4000 characters should be enough for everybody
			http.Error(w, "Text too long ("+strconv.Itoa(maxTextLength)+" characters max)", http.StatusBadRequest)
			return
		}

		pinStr := GetPin()
		if pinStr == "0" {
			//no pin could be generated
			fmt.Println("Error: PIN could not be generated")
			w.WriteHeader(http.StatusServiceUnavailable) //http 503
			return
		}

		m[pinStr] = string

		time.AfterFunc(time.Minute*5, func() {
			if m[pinStr] != "" {
				log.Println("Auto-deleted text for PIN")
				delete(m, pinStr)
			}
		})

		//fmt.Println("Returning created PIN " + sPin)
		io.WriteString(w, pinStr)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/pin", func(w http.ResponseWriter, r *http.Request) {
		initResult := InitRequest(w, r)
		if !initResult {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError) //http 500
			return
		}

		dt := time.Now()
		blen := strconv.Itoa(len(body))
		fmt.Println("Received request to retrieve text with PIN of length "+blen+" at ", dt.String())
		pinStr := string(body)
		//TODO: make environment-specific logging for debugging purposes
		//fmt.Println("Received request POST /pin with body " + pinStr)

		pwd := m[pinStr]
		if pwd == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		delete(m, pinStr)
		//TODO: make environment-specific logging for debugging purposes
		//fmt.Println("Returning retrieved string " + pwd)
		io.WriteString(w, pwd)
	}).Methods("POST", "OPTIONS")

	sttport, sttportExists := os.LookupEnv("STT_PORT")
	if !sttportExists {
		http.ListenAndServe(":9999", r)
	} else {
		http.ListenAndServe(":"+sttport, r)
	}
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
or 0 if no PIN could be generated due to collisions
*/
func GetPin() string {
	var src cryptoSource
	rnd := rand.New(src)

	pin := uint64(rnd.Intn(99999))
	pinStr := strconv.FormatUint(pin, 10)
	var i int = 0
	var numberOfTimesToTry = 100 //100 subsequent collisions should be very unlikely
	valuePresent := false
	keepTrying := true
	for ok := true; ok; ok = keepTrying {
		_, valuePresent = m[pinStr]
		keepTrying = valuePresent && i < numberOfTimesToTry
		i++
	}

	if !valuePresent {
		return pinStr
	}

	return "0"
}

/*
InitRequest returns true if the initialization succeeded, otherwise
false. Should be used to in the beginning of outside-facing
request methods.
*/
func InitRequest(w http.ResponseWriter, r *http.Request) bool {
	sttenv, exists := os.LookupEnv("STT_ENV")

	if exists && sttenv == "dev" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return false
		}
	}

	xReqW := r.Header.Get("X-Requested-With")

	if xReqW != "XMLHttpRequest" {
		http.Error(w, "Bad header", http.StatusBadRequest)
		return false
	}

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	return true
}
