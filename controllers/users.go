package controllers

import (
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/views"
)

// NewUsers ...
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "users/new"),
		us:      us,
	}
}

// Users ...
type Users struct {
	NewView *views.View
	us      *models.UserService
}

// New ...
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	err := u.NewView.Render(w, nil)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}

// SignupForm ...
type SignupForm struct {
	FullName     string `schema:"name"`
	EmailAddress string `schema:"email"`
	Password     string `schema:"password"`
}

// Create ...
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var signupForm SignupForm
	if err := parseForm(r, &signupForm); err != nil {
		log.Fatalln("ERROR:", err)
	}

	user := models.User{
		Name:         signupForm.FullName,
		EmailAddress: signupForm.EmailAddress,
	}

	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Fprintln(w, signupForm)
	http.Redirect(w, r, "/", http.StatusFound)
}
