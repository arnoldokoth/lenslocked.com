package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/rand"
	"github.com/arnoldokoth/lenslocked.com/views"
)

// ErrGeneric rendered when something goes wrong and we
// ain't got nothing better to tell the user
var ErrGeneric = errors.New("Oops... Something Went Wrong")

// NewUsers ...
func NewUsers(us models.UserService) *Users {
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
	us        models.UserService
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
	var vd views.Data
	var signupForm SignupForm
	if err := parseForm(r, &signupForm); err != nil {
		log.Println("u.Create() ERROR:", err)
		vd.SetAlert(err)
		u.NewView.Render(w, vd)
		return
	}

	user := models.User{
		Name:         signupForm.FullName,
		EmailAddress: signupForm.EmailAddress,
		Password:     signupForm.Password,
	}

	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, vd)
		return
	}

	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// LoginForm ...
type LoginForm struct {
	EmailAddress string `schema:"email"`
	Password     string `schema:"password"`
}

// Login ...
// GET & POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var loginForm LoginForm
	if err := parseForm(r, &loginForm); err != nil {
		http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
		return
	}

	user, err := u.us.Authenticate(loginForm.EmailAddress, loginForm.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			fmt.Fprintln(w, "Invalid Email Address")
		case models.ErrInvalidPassword:
			fmt.Fprintln(w, "Invalid Password Provided")
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = u.signIn(w, user)
	if err != nil {
		http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	return nil
}

// CookieTest ...
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, user)
}
