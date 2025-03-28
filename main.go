package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var pointCache = make(map[uuid.UUID]int)

func main() {
    
    var use_logging bool
    flag.BoolVar(&use_logging, "logging", false, "enable logging for all requests")
    var port int
    flag.IntVar(&port, "port", 8080, "port used for the web server (uint16)")
    if port < 1 || port > 65535 {
        log.Fatal("specified port number must be between 0 and 65535 (uint16)")
    }
    flag.Parse()
    
    r := mux.NewRouter()
    r.HandleFunc("/receipts/process", process).
      Methods("POST") //.
      //Headers("Content-Type", "application/json")
    r.HandleFunc("/receipts/{id}/points", getPoints).
      Methods("GET")

    if use_logging {
        log.Println("using logging")
        http.Handle("/", logRequestHandler(r))
    } else {
        http.Handle("/", r)
    }

    log.Println("Starting on port: ", port)
    log.Fatal(http.ListenAndServe(fmt.Sprint(":", port), nil))
}

func process(w http.ResponseWriter, r *http.Request) {
    invalid := func() {
        w.WriteHeader(400)
        w.Write([]byte("The receipt is invalid"))
    }

    body, err := io.ReadAll(r.Body)
    if err != nil { 
        invalid() 
        return
    }

    receipt := Receipt{}

    json.Unmarshal(body, &receipt)

    valid := receipt.verify()
    if !valid {
        invalid()
        return
    }

    uuid := uuid.NewSHA1(uuid.Max, body)

    response := struct {
        Id string `json:"id"`
    }{
        uuid.String(),
    }

    json, err := json.Marshal(response)
    if err != nil {
        invalid()
        return
    }

    if _, exists := pointCache[uuid]; !exists {
        p := receipt.calculatePoints()
        pointCache[uuid] = p
    }

    w.Write(json)
}

func getPoints(w http.ResponseWriter, r *http.Request) {
    invalid := func() {
        w.WriteHeader(404)
        w.Write([]byte("No receipt found for that id"))
    }

    vars := mux.Vars(r)

    id := vars["id"]

    uuid, err := uuid.Parse(id)
    if err != nil {
        invalid()
        return
    }

    p, ok := pointCache[uuid]
    if !ok {
        invalid()
        return
    }

    response := struct {
        Points int64 `json:"points"`
    }{
        int64(p),
    }

    json, err := json.Marshal(response)
    if err != nil {
        invalid()
        return
    }

    w.Write(json)
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		uri := r.URL.String()
		method := r.Method
        headers := r.Header
        headerBuilder := strings.Builder{}
        headerBuilder.Write([]byte("headers: "))
        for key, value := range headers {
            headerBuilder.Write([]byte(fmt.Sprint("'", key, " : ", value, "', ")))
        }
        logHTTPReq(uri, method, headerBuilder.String())
	}

	return http.HandlerFunc(fn)
}

func logHTTPReq(s ...string) {
    log.Println("New request received:")
    for _, v := range s {
        if strings.TrimSpace(v) == "" { continue }
        log.Println(v)
    }
}
