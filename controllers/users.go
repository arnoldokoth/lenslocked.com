package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arnoldokoth/lenslocked.com/context"
	"github.com/arnoldokoth/lenslocked.com/email"
	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/rand"
	"github.com/arnoldokoth/lenslocked.com/views"
)

// ErrGeneric rendered when something goes wrong and we
// ain't got nothing better to tell the user
var ErrGeneric = errors.New("Oops... Something Went Wrong")

// NewUsers ...
func NewUsers(us models.UserService, emailer *email.Client) *Users {
	return &Users{
		CreateView: views.NewView("bootstrap", "users/new"),
		LoginView:  views.NewView("bootstrap", "users/login"),
		us:         us,
		emailer:    emailer,
	}
}

// Users ...
type Users struct {
	CreateView *views.View
	LoginView  *views.View
	us         models.UserService
	emailer    *email.Client
}

// SignupForm ...
type SignupForm struct {
	FullName     string `schema:"name"`
	EmailAddress string `schema:"email"`
	Password     string `schema:"password"`
}

// New ...
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var signupForm SignupForm
	parseURLParams(r, &signupForm)
	u.CreateView.Render(w, r, signupForm)
}

// Create ...
// GET & POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var signupForm SignupForm
	vd.Yield = &signupForm
	if err := parseForm(r, &signupForm); err != nil {
		log.Println("users.Create() ERROR:", err)
		vd.SetAlert(err)
		u.CreateView.Render(w, r, vd)
		return
	}

	user := models.User{
		Name:         signupForm.FullName,
		EmailAddress: signupForm.EmailAddress,
		Password:     signupForm.Password,
	}

	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.CreateView.Render(w, r, vd)
		return
	}

	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	go u.emailer.Welcome(user.Name, user.EmailAddress)

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Welcome To LensLocked.com!",
	}

	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
}

// LoginForm ...
type LoginForm struct {
	EmailAddress string `schema:"email"`
	Password     string `schema:"password"`
}

// Login ...
// GET & POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var loginForm LoginForm
	if err := parseForm(r, &loginForm); err != nil {
		log.Println("u.Login() ERROR:", err)
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	user, err := u.us.Authenticate(loginForm.EmailAddress, loginForm.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("Invalid Email Address")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: fmt.Sprintf("Welcome Back %v!", user.Name),
	}

	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
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

// Logout ...
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	user := context.User(r.Context())
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.Update(user)

	http.Redirect(w, r, "/", http.StatusFound)
}
