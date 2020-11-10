package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/controllers"
	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/gorilla/mux"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "arnoldokoth"
	password = "Password123!"
	dbname   = "lenslocked_dev"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	userService, err := models.NewUserService(psqlInfo)
	must(err)

	defer userService.Close()

	must(userService.AutoMigrate())

	router := mux.NewRouter()

	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(userService)

	// Static Routes
	router.Handle("/", staticController.Home).Methods("GET")
	router.Handle("/contact", staticController.Contact).Methods("GET")
	router.Handle("/faq", staticController.FAQ).Methods("GET")

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
