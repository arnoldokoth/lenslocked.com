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
	services, err := models.NewServices(psqlInfo)
	must(err)

	defer services.Close()

	must(services.AutoMigrate())

	router := mux.NewRouter()

	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery)

	// Static Routes
	router.Handle("/", staticController.Home).Methods("GET")
	router.Handle("/contact", staticController.Contact).Methods("GET")
	router.Handle("/faq", staticController.FAQ).Methods("GET")

	// User Routes
	router.Handle("/signup", usersController.NewView).Methods("GET")
	router.HandleFunc("/signup", usersController.Create).Methods("POST")
	router.Handle("/login", usersController.LoginView).Methods("GET")
	router.HandleFunc("/login", usersController.Login).Methods("POST")
	router.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")

	// Gallery Routes
	router.HandleFunc("/galleries/new", galleriesController.New).Methods("GET")
	router.HandleFunc("/galleries/new", galleriesController.Create).Methods("POST")

	log.Printf("Server Running On Port: %d", 3000)
	http.ListenAndServe(":3000", router)
}

func must(err error) {
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}
