package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/gorilla/schema"
	"github.com/seehuhn/mt19937"
)

// RandomNumber user defined type
// HandleRandom returns this
type RandomNumber struct {
	Number int64
}

// NewRandomNumber create random number using mersenne twister PRNG
func NewRandomNumber() *RandomNumber {
	var rng = rand.New(mt19937.New())
	rn := RandomNumber{}
	rng.Seed(time.Now().UnixNano())
	rn.Number = rng.Int63()
	return &rn
}

// RandomRequest Model for QueryString parameters
type RandomRequest struct {
	Start int
	End   int
}

// Decoder the decodes the query string parameters
var decoder = schema.NewDecoder()

// Count the number of random requests
var requestCount int64 = 0

func HandleRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Println("Invalid method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var start, end = 0, 0
	var err error = nil

	var randomRequest = RandomRequest{}

	err = r.ParseForm()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = decoder.Decode(&randomRequest, r.Form)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	start, end = randomRequest.Start, randomRequest.End

	if start == 0 || end == 0 {
		log.Println("Start or end 0")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if start < 0 || end < 0 {
		log.Println("Start or end < 0")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if start > end {
		log.Println("Start > End")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rn := NewRandomNumber()
	rn.Number = rn.Number%(int64(end)-int64(start+1)) + int64(start)
	b, err := json.Marshal(rn)

	if err != nil {
		log.Println("Can't validate JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s", string(b))

	requestCount++
}

var healthmx sync.Mutex

func HandleHealth(w http.ResponseWriter, r *http.Request) {

	healthmx.Lock()
	defer healthmx.Unlock()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, `{ "status": "ok", "requests": %d }`, requestCount)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	case "/health":
		HandleHealth(w, r)
	case "/random":
		HandleRandom(w, r)
	default:
		log.Println(path)
		w.WriteHeader(http.StatusNotFound)
	}
}

type Server struct {
}

func main() {

	s := &http.Server{
		Addr:           ":8080",
		Handler:        &Server{},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())

}
