package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// The person Type (more like an object)
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
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

	for i := 0; i < 5000; i++ {
		f := fnames[rand.Intn(10)]
		l := lnames[rand.Intn(10)]
		c := cities[rand.Intn(10)]
		s := states[rand.Intn(10)]
		people = append(people, Person{ID: strconv.Itoa(i + 1), Firstname: f, Lastname: l, Address: &Address{City: c, State: s}})
	}
}

// our main function
func main() {
	router := mux.NewRouter()

	// initial data population
	PopulateInitialData()

	//Add endpoints
	router.HandleFunc("/people", GetPeople).Methods("GET")
	router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	router.HandleFunc("/people/{id}", CreatePerson).Methods("POST")
	router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")

	router.HandleFunc("/throw", ThrowError).Methods("GET")
	router.HandleFunc("/slowproc/{num}", GetDataSlowly).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

//Throw an error
func ThrowError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "You Can't Do That!", http.StatusTeapot)
	return
}

//Artificially long running process
func GetDataSlowly(w http.ResponseWriter, r *http.Request) {
	chars := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	s := ""
	params := mux.Vars(r)
	num := params["num"]
	ctr, err := strconv.Atoi(num)
	if err != nil {
		ctr = 10
	}
	if ctr > 250 {
		ctr = 250
	}

	s2 := rand.NewSource(time.Now().UnixNano())
	r2 := rand.New(s2)

	for i := 0; i < 1024*ctr; i++ {
		//s += "c"
		s += chars[r2.Intn(36)]
	}
	json.NewEncoder(w).Encode(s)
}

//Create blank endpoint functions (for now)
// Display all from the people var
func GetPeople(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(people)
}

// Display a single data
func GetPerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, item := range people {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Person{})
}

// create a new item
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var person Person
	_ = json.NewDecoder(r.Body).Decode(&person)
	person.ID = params["id"]
	people = append(people, person)
	json.NewEncoder(w).Encode(people)
}

// Delete an item
func DeletePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for index, item := range people {
		if item.ID == params["id"] {
			people = append(people[:index], people[index+1:]...)
			break
		}
		json.NewEncoder(w).Encode(people)
	}
}
