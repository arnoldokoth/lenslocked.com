package controllers

import "github.com/arnoldokoth/lenslocked.com/views"

// NewStatic ...
func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "static/home"),
		FAQ:     views.NewView("bootstrap", "static/faq"),
		Contact: views.NewView("bootstrap", "static/contact"),
	}
}

// Static ...
type Static struct {
	Home    *views.View
	FAQ     *views.View
	Contact *views.View
}
