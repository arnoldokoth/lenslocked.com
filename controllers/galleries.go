package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arnoldokoth/lenslocked.com/context"
	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/views"
	"github.com/gorilla/mux"
)

const (
	showGallery = "show_gallery"
)

// NewGalleries ...
func NewGalleries(gs models.GalleryService, router *mux.Router) *Galleries {
	return &Galleries{
		CreateView: views.NewView("bootstrap", "galleries/new"),
		ShowView:   views.NewView("bootstrap", "galleries/show"),
		gs:         gs,
		router:     router,
	}
}

// Galleries ...
type Galleries struct {
	CreateView *views.View
	ShowView   *views.View
	gs         models.GalleryService
	router     *mux.Router
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
// POST /galleries/new
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var galleryForm GalleryForm
	if err := parseForm(r, &galleryForm); err != nil {
		log.Println("galleries.Create() ERROR:", err)
		vd.SetAlert(err)
		g.CreateView.Render(w, vd)
		return
	}

	user := context.User(r.Context())

	gallery := models.Gallery{
		Title:  galleryForm.Title,
		UserID: user.ID,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.CreateView.Render(w, vd)
		return
	}

	url, err := g.router.Get(showGallery).URL("id", fmt.Sprintf("%+v", gallery.ID))
	if err != nil {
		// TODO: Redirect To Galleries Index Page
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

// Show ,,,
// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid Gallery ID", http.StatusNotFound)
		return
	}

	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery Not Found", http.StatusNotFound)
		default:
			http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
		}
		return
	}

	vd.Yield = gallery

	g.ShowView.Render(w, vd)
}
