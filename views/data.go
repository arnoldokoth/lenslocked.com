package views

import (
	"net/http"
	"time"

	"github.com/arnoldokoth/lenslocked.com/models"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
	AlertMsgGeneric = "Ooops... Something Went Wrong!"
)

// Alert ...
type Alert struct {
	Level   string
	Message string
}

// Data ...
type Data struct {
	Alert *Alert
	User  *models.User
	Yield interface{}
}

// SetAlert ...
func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: pErr.Public(),
		}
	} else {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: AlertMsgGeneric,
		}
	}
}

// AlertError ...
func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(5 * time.Minute)
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	msg := http.Cookie{
		Name:     "alert_message",
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}

	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func clearAlert(w http.ResponseWriter) {
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	msg := http.Cookie{
		Name:     "alert_message",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}

	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func getAlert(r *http.Request) *Alert {
	lvl, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}

	msg, err := r.Cookie("alert_message")
	if err != nil {
		return nil
	}

	alert := Alert{
		Level:   lvl.Value,
		Message: msg.Value,
	}

	return &alert
}

// RedirectAlert ...
func RedirectAlert(w http.ResponseWriter, r *http.Request, urlStr string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlStr, code)
}
