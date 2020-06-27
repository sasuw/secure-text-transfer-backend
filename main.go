package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var m map[uint64]string

func main() {

	m = make(map[uint64]string)

	r := mux.NewRouter()

	r.HandleFunc("/string/{string}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		string := vars["string"]

		pin := uint64(rand.Intn(10000))
		m[pin] = string

		fmt.Fprintf(w, "You have posted string: %s\n", string)

		sPin := strconv.FormatUint(pin, 32)
		w.Write([]byte(sPin))
	}).Methods("POST")

	r.HandleFunc("/pin/{pin}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pin := vars["pin"]

		fmt.Fprintf(w, "You have posted PIN: %s\n", pin)
	}).Methods("GET")

	http.ListenAndServe(":10000", r)
}
