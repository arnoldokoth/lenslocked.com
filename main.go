package main

import (
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/controllers"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers()

	// Static Routes
	router.Handle("/", staticController.HomeView).Methods("GET")
	router.Handle("/contact", staticController.ContactView).Methods("GET")
	router.Handle("/faq", staticController.FAQView).Methods("GET")

	// User Routes
	router.HandleFunc("/signup", usersController.New).Methods("GET")
	router.HandleFunc("/signup", usersController.Create).Methods("POST")

	log.Printf("Server Running On Port: %d", 3000)
	http.ListenAndServe(":3000", router)
}

func must(err error) {
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}
