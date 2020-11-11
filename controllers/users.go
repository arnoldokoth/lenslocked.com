package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/views"
)

// ErrGeneric rendered when something goes wrong and we
// ain't got nothing better to tell the user
var ErrGeneric = errors.New("Ooops... Something Went Wrong")

// NewUsers ...
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

// Users ...
type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        *models.UserService
}

// SignupForm ...
type SignupForm struct {
	FullName     string `schema:"name"`
	EmailAddress string `schema:"email"`
	Password     string `schema:"password"`
}

// Create ...
// GET & POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := u.NewView.Render(w, nil)
		if err != nil {
			http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var signupForm SignupForm
		if err := parseForm(r, &signupForm); err != nil {
			log.Fatalln("ERROR:", err)
		}

		user := models.User{
			Name:         signupForm.FullName,
			EmailAddress: signupForm.EmailAddress,
		}

		if err := u.us.Create(&user); err != nil {
			http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
			return
		}

		log.Println(user)
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// LoginForm ...
type LoginForm struct {
	EmailAddress string `schema:"email"`
	Password     string `schema:"password"`
}

// Login ...
// GET & POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := u.LoginView.Render(w, nil)
		if err != nil {
			http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var loginForm LoginForm
		if err := parseForm(r, &loginForm); err != nil {
			http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, loginForm)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
