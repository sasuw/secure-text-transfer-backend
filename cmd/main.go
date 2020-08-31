package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
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
	"golang.org/x/time/rate"
)

/*
Payload represents data sent from client when storing client-sidedly encrypted text
*/
type Payload struct {
	Ciphertext string
	Iv         string
	Salt       string
	Id         string
}

var m map[string]string
var mp map[string]Payload
var limiter = rate.NewLimiter(1, 3)

func main() {

	m = make(map[string]string)
	mp = make(map[string]Payload)

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		initResult := InitRequest(w, r)
		if !initResult {
			return
		}

		dt := time.Now()
		fmt.Println("Received request to get status at ", dt.String())

		w.WriteHeader(http.StatusNoContent)
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/encryptedText", func(w http.ResponseWriter, r *http.Request) {
		initResult := InitRequest(w, r)
		if !initResult {
			return
		}

		dt := time.Now()
		fmt.Println("Received request to store json at ", dt.String())

		var p Payload

		err := decodeJSONBody(w, r, &p)
		if err != nil {
			var mr *malformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.msg, mr.status)
			} else {
				log.Println(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		if p.Ciphertext == "" || p.Id == "" || p.Iv == "" || p.Salt == "" {
			log.Println("Required payload field missing")
			w.WriteHeader(http.StatusBadRequest) //http 400
			return
		}

		mp[p.Id] = p

		time.AfterFunc(time.Minute*5, func() {
			if mp[p.Id] != (Payload{}) {
				log.Println("Auto-deleted Payload")
				delete(mp, p.Id)
			}
		})

		w.WriteHeader(http.StatusNoContent) //http 204
		return
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/encryptedText", func(w http.ResponseWriter, r *http.Request) {
		initResult := InitRequest(w, r)
		if !initResult {
			return
		}

		id, _ := r.URL.Query()["id"]

		if len(id) > 1 {
			log.Fatal("Multiple IDs not allowed")
			w.WriteHeader(http.StatusBadRequest) //http 400
			return
		}
		idStr := id[0]

		dt := time.Now()
		blen := strconv.Itoa(len(idStr))
		fmt.Println("Received request to retrieve encrypted text with passphrase of length "+blen+" at ", dt.String())
		//TODO: make environment-specific logging for debugging purposes
		//fmt.Println("Received request POST /pin with body " + pinStr)

		p := mp[idStr]
		if p == (Payload{}) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		delete(mp, idStr)

		pJson, err := json.Marshal(p)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError) //http 500
			return
		}

		//TODO: make environment-specific logging for debugging purposes
		pJsonStr := string(pJson)
		fmt.Println("Returning retrieved struct " + pJsonStr)

		io.WriteString(w, pJsonStr)
	}).Methods("GET")

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
			http.Error(w, "Text too long ("+strconv.Itoa(maxTextLength)+" characters max)", http.StatusBadRequest) //400
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
		http.ListenAndServe(":9999", limit(r))
	} else {
		http.ListenAndServe(":"+sttport, limit(r))
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
GetPin returns a PIN of type string with a value between "1" and "99999"
or "0" if no PIN could be generated due to collisions
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
		http.Error(w, "Bad header", http.StatusBadRequest) //400
		return false
	}

	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	return true
}
