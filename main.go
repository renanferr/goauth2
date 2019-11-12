package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func loggerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("<< %s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	responseType := params["response_type"]
	// clientID := params["client_id"]
	// redirectURI := params["redirect_uri"]
	// scope := params["read"]

	io.WriteString(w, strings.Join(responseType, ","))
}

type TokenResponse struct {
	clientID     string
	clientSecret string
	grantType    string
	code         string
	redirectURI  string
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	clientID := params["client_id"]
	clientSecret := params["client_secret"]
	grantType := params["grant_type"]
	code := params["code"]
	redirectURI := params["redirect_uri"]
	w.Header().Set("Content-Type", "application/json")
	response := TokenResponse{
		clientID:     clientID[0],
		clientSecret: clientSecret[0],
		grantType:    grantType[0],
		code:         code[0],
		redirectURI:  redirectURI[0],
	}
	fmt.Printf("%v", response)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Println(responseJSON)
	w.Write(responseJSON)
}

func main() {
	errorChain := alice.New(loggerHandler, recoverHandler)

	r := mux.NewRouter()

	r.HandleFunc("/authorize", AuthorizeHandler).Methods("GET")

	r.HandleFunc("/oauth/token", TokenHandler).Methods("GET")

	http.Handle("/", errorChain.Then(r))

	server := &http.Server{
		Addr: ":8888",
	}

	err := server.ListenAndServe()

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
