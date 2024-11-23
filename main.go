package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)


func getIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Hello, World!"}`)
}

func bookIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	books := []map[string]string{
		{"id": "1", "title": "Book 1"},
		{"id": "2", "title": "Book 2"},
		{"id": "3", "title": "Book 3"},
	}

	json.NewEncoder(w).Encode(books)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", getIndex).Methods("GET")

	bookRouter := r.PathPrefix("/books").Subrouter()
	bookRouter.HandleFunc("", bookIndex).Methods("GET")

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}