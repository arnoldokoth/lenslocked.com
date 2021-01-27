package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arnoldokoth/lenslocked.com/controllers"
	"github.com/arnoldokoth/lenslocked.com/email"
	"github.com/arnoldokoth/lenslocked.com/middleware"
	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/rand"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

func main() {
	cfg := LoadConfig()
	dbConfig := cfg.Database
	services, err := models.NewServices(
		models.WithGorm(dbConfig.Dialect(), dbConfig.ConnString()),
		models.WithLogMode(!cfg.IsProd()),
		models.WithUser(cfg.HMACKey, cfg.Pepper),
		models.WithGallery(),
		models.WithImage(),
	)
	must(err)

	defer services.Close()

	must(services.AutoMigrate())

	mgCfg := cfg.Mailgun
	emailer := email.NewClient(
		email.WithMailgun(mgCfg.Domain, mgCfg.APIKey, mgCfg.PublicAPIKey),
		email.WithSender("Lenslocked.com Support", fmt.Sprintf("support@%s", mgCfg.Domain)),
	)

	router := mux.NewRouter()

	dbxOAuth := &oauth2.Config{
		ClientID:     cfg.Dropbox.ID,
		ClientSecret: cfg.Dropbox.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.Dropbox.AuthURL,
			TokenURL: cfg.Dropbox.TokenURL,
		},
		RedirectURL: "http://localhost:3000/oauth/dropbox/callback",
	}

	userMw := middleware.User{UserService: services.User}
	requireUserMw := middleware.RequireUser{User: userMw}

	bytes, _ := rand.Bytes(32)
	csrfMw := csrf.Protect(bytes, csrf.Secure(cfg.IsProd()))

	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User, emailer)
	galleriesController := controllers.NewGalleries(services.Gallery, services.Image, router)

	dbxRedirect := func(w http.ResponseWriter, r *http.Request) {
		state := csrf.Token(r)
		cookie := http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		fmt.Println("State:", state)
		url := dbxOAuth.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	}

	dbxCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.FormValue("state")
		cookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, "Invalid State", http.StatusBadRequest)
			return
		} else if cookie == nil || cookie.Value != state {
			http.Error(w, "Invalid State Provided", http.StatusBadRequest)
			return
		}
		cookie.Value = ""
		cookie.Expires = time.Now()
		http.SetCookie(w, cookie)

		code := r.FormValue("code")
		token, err := dbxOAuth.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%+v", token)
	}

	router.HandleFunc("/oauth/dropbox/callback", dbxCallback)

	router.HandleFunc("/oauth/dropbox/connect", requireUserMw.ApplyFn(dbxRedirect)).Methods("GET")

	// Static Routes
	router.Handle("/", staticController.Home).Methods("GET")
	router.Handle("/contact", staticController.Contact).Methods("GET")
	router.Handle("/faq", staticController.FAQ).Methods("GET")

	assetHandler := http.FileServer(http.Dir("./assets"))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetHandler))

	// User Routes
	router.HandleFunc("/signup", usersController.New).Methods("GET")
	router.HandleFunc("/signup", usersController.Create).Methods("POST")
	router.Handle("/login", usersController.LoginView).Methods("GET")
	router.HandleFunc("/login", usersController.Login).Methods("POST")
	router.HandleFunc("/logout", requireUserMw.ApplyFn(usersController.Logout)).Methods("POST")

	// Gallery Routes
	router.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesController.Index)).Methods("GET")
	router.Handle("/galleries/new", requireUserMw.Apply(galleriesController.CreateView)).Methods("GET")
	router.HandleFunc("/galleries/new", requireUserMw.ApplyFn(galleriesController.Create)).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}", requireUserMw.ApplyFn(galleriesController.Show)).Methods("GET").Name("show_gallery")
	router.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesController.Edit)).Methods("GET").Name("edit_gallery")
	router.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesController.Update)).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleriesController.Delete)).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleriesController.Upload)).Methods("POST")
	router.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFn(galleriesController.ImageDelete)).Methods("POST")

	// Image Routes
	imageHandler := http.FileServer(http.Dir("./images"))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	log.Printf("Server Running On Port: %d", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMw.Apply(router)))
}

func must(err error) {
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}
