package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

// The person Type (more like an object)
type Person struct {
	ID        string   `json:"id"`
	Firstname string   `json:"firstname"`
	Lastname  string   `json:"lastname"`
	Address   *Address `json:"address,omitempty"`
}
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people []Person

//Populate initial data
func PopulateInitialData() {
	fnames := []string{"John", "Paul", "George", "Steve", "David", "Chris", "Dawn", "Sarah", "Amy", "Dena"}
	lnames := [10]string{"Jones", "Smith", "Johnson", "Mills", "Elliot", "Meyers", "Nelson", "Hayes", "Pollard", "Farmer"}
	cities := [10]string{"Monroe", "Rockrord", "Springfield", "Franklin", "Washington", "Salem", "Fairview", "Bristol", "Madison", "Georgetown"}
	states := [10]string{"IL", "WI", "AL", "MT", "CA", "WA", "OR", "VT", "FL", "TX"}

	for i := 0; i < 1000; i++ {
		f := fnames[rand.Intn(10)]
		l := lnames[rand.Intn(10)]
		c := cities[rand.Intn(10)]
		s := states[rand.Intn(10)]
		people = append(people, Person{ID: strconv.Itoa(i + 1), Firstname: f, Lastname: l, Address: &Address{City: c, State: s}})
	}
}

// our main function
// func main() {
// 	router := mux.NewRouter()

// 	// initial data population
// 	PopulateInitialData()

// 	//Add endpoints
// 	router.HandleFunc("/people", GetPeople).Methods("GET", "OPTIONS")
// 	router.HandleFunc("/people/{id}", GetPerson).Methods("GET", "OPTIONS")
// 	router.HandleFunc("/people/{id}", CreatePerson).Methods("POST", "OPTIONS")
// 	router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE", "OPTIONS")

// 	router.HandleFunc("/throw", ThrowError).Methods("GET", "OPTIONS")
// 	router.HandleFunc("/hang", HangMe).Methods("GET", "OPTIONS")
// 	router.HandleFunc("/slowproc/{num}", GetDataSlowly).Methods("GET", "OPTIONS")
// 	router.HandleFunc("/crash", CrashIt).Methods("GET", "OPTIONS")
// 	log.Fatal(http.ListenAndServe(":8000", router))
// }

type key int

const (
	requestIDKey key = 0
)

var (
	listenAddr string
	healthy    int32
)

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":8000", "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := mux.NewRouter()

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	var srv = &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// initial data population
	PopulateInitialData()

	//Add endpoints
	router.HandleFunc("/people", GetPeople).Methods("GET", "OPTIONS")
	router.HandleFunc("/people/{id}", GetPerson).Methods("GET", "OPTIONS")
	router.HandleFunc("/people/{id}", CreatePerson).Methods("POST", "OPTIONS")
	router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE", "OPTIONS")

	router.HandleFunc("/throw", ThrowError).Methods("GET", "OPTIONS")
	router.HandleFunc("/hang", HangMe).Methods("GET", "OPTIONS")
	router.HandleFunc("/slowproc/{num}", GetDataSlowly).Methods("GET", "OPTIONS")
	router.HandleFunc("/shutdown", KillIt).Methods("GET", "OPTIONS")

	// idleConnsClosed := make(chan struct{})
	// go func() {
	// 	sigint := make(chan os.Signal, 1)
	// 	signal.Notify(sigint, os.Interrupt)
	// 	<-sigint

	// 	// We received an interrupt signal, shut down.
	// 	if err := srv.Shutdown(context.Background()); err != nil {
	// 		// Error from closing listeners, or context timeout:
	// 		log.Printf("HTTP server Shutdown: %v", err)
	// 	}
	// 	close(idleConnsClosed)
	// }()

	// if err := srv.ListenAndServe(); err != http.ErrServerClosed {
	// 	// Error starting or closing listener:
	// 	log.Printf("HTTP server ListenAndServe: %v", err)
	// }

	// <-idleConnsClosed

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")

}

//Kill the app
func KillIt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server should be shutting down...")

	os.Exit(0)
	//panic("Oh, I'm Dying!")
	return
}

//Hang on this call
func HangMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ch := make(chan int)
	ch <- 5
	return
}

//Throw an error
func ThrowError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	whch := rand.Intn(3)

	if whch == 0 {
		http.Error(w, "{ \"status\" : 405,  \"userMessage\" : \"You can't do that!\", \"errorUrl\" : \"https://httpstatusdogs.com/img/405.jpg\" }", http.StatusMethodNotAllowed)
	}

	if whch == 1 {
		http.Error(w, "{ \"status\" : 403, \"userMessage\" : \"Forbidden!\", \"errorUrl\" : \"https://httpstatusdogs.com/img/403.jpg\" }", http.StatusForbidden)
	}
	if whch == 2 {
		http.Error(w, "{ \"status\" : 414, \"userMessage\" : \"Whoa Big Fella!\", \"errorUrl\" : \"https://httpstatusdogs.com/img/414.jpg\" }", http.StatusRequestURITooLong)
	}
	if whch == 3 {
		http.Error(w, "Payment Required", http.StatusPaymentRequired)
	}
	if whch == 4 {
		http.Error(w, "Gone", http.StatusGone)
	}
	if whch == 5 {
		http.Error(w, "Length Required", http.StatusLengthRequired)
	}
	if whch == 6 {
		http.Error(w, "Locked", http.StatusLocked)
	}
	if whch == 7 {
		http.Error(w, "You Can't Do That!", http.StatusTeapot)
	}
	
	return
}

//Artificially long running process
func GetDataSlowly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	chars := []string{" ", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	s := ""
	params := mux.Vars(r)
	num := params["num"]
	ctr, err := strconv.Atoi(num)
	if err != nil {
		ctr = 10
	}
	if ctr > 1000 {
		ctr = 1000
	}

	s2 := rand.NewSource(time.Now().UnixNano())
	r2 := rand.New(s2)

	for i := 0; i < 1024*ctr; i++ {
		//s += "c"
		s += chars[r2.Intn(37)]
	}
	json.NewEncoder(w).Encode(s)
}

//Create blank endpoint functions (for now)
// Display all from the people var
func GetPeople(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//PopulateInitialData()
	json.NewEncoder(w).Encode(people)
}

// Display a single data
func GetPerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//PopulateInitialData()
	params := mux.Vars(r)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(os.Stdout).Encode(item)
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

// create a new item
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	params := mux.Vars(r)
	var person Person
	_ = json.NewDecoder(r.Body).Decode(&person)
	person.ID = params["id"]
	people = append(people, person)
	json.NewEncoder(w).Encode(people)
}

// Delete an item
func DeletePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	params := mux.Vars(r)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
			break
		}
		json.NewEncoder(w).Encode(people)
	}
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
