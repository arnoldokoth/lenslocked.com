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
		NewView: views.NewView("bootstrap", "galleries/new"),
		gs:      gs,
	}
}

// Galleries ...
type Galleries struct {
	NewView *views.View
	gs      models.GalleryService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// New ...
func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.NewView.Render(w, nil)
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var galleryForm GalleryForm
	if err := parseForm(r, &galleryForm); err != nil {
		log.Println("galleries.Create() ERROR:", err)
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		UserID: 1,
		Title:  galleryForm.Title,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
