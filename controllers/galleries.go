package controllers

import (
	"log"
	"net/http"

	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/views"
)

// NewGalleries ...
func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		CreateView: views.NewView("bootstrap", "galleries/new"),
		gs:         gs,
	}
}

// Galleries ...
type Galleries struct {
	CreateView *views.View
	gs         models.GalleryService
}

// GalleryForm ...
type GalleryForm struct {
	Title string `schema:"title"`
}

// New ...
func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.CreateView.Render(w, nil)
}

// Create ...
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var galleryForm GalleryForm
	if err := parseForm(r, &galleryForm); err != nil {
		log.Println("galleries.Create() ERROR:", err)
		vd.SetAlert(err)
		g.CreateView.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		Title:  galleryForm.Title,
		UserID: 1,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.CreateView.Render(w, vd)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
