package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/controllers"
	"github.com/arnoldokoth/lenslocked.com/middleware"
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

	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery, router)

	// Static Routes
	router.Handle("/", staticController.Home).Methods("GET")
	router.Handle("/contact", staticController.Contact).Methods("GET")
	router.Handle("/faq", staticController.FAQ).Methods("GET")

	// User Routes
	router.Handle("/signup", usersController.CreateView).Methods("GET")
	router.HandleFunc("/signup", usersController.Create).Methods("POST")
	router.Handle("/login", usersController.LoginView).Methods("GET")
	router.HandleFunc("/login", usersController.Login).Methods("POST")

	// Testing Routes
	router.HandleFunc("/cookietest", requireUserMw.ApplyFn(usersController.CookieTest)).Methods("GET")

	// Gallery Routes
	router.Handle("/galleries/new", requireUserMw.Apply(galleriesController.CreateView)).Methods("GET")
	router.HandleFunc("/galleries/new", requireUserMw.ApplyFn(galleriesController.Create)).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}", requireUserMw.ApplyFn(galleriesController.Show)).Methods("GET").Name("show_gallery")
	router.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesController.Edit)).Methods("GET").Name("edit_gallery")
	router.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesController.Update)).Methods("POST")

	log.Printf("Server Running On Port: %d", 3000)
	http.ListenAndServe(":3000", router)
}

func must(err error) {
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}
