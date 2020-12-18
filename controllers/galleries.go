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
		EditView:   views.NewView("bootstrap", "galleries/edit"),
		gs:         gs,
		router:     router,
	}
}

// Galleries ...
type Galleries struct {
	CreateView *views.View
	ShowView   *views.View
	EditView   *views.View
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

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid Gallery ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery Not Found", http.StatusNotFound)
		default:
			http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
		}
		return nil, err
	}

	return gallery, nil
}

// Show ,,,
// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	vd.Yield = gallery

	g.ShowView.Render(w, vd)
}

// Edit ,,,
// GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}

	vd.Yield = gallery
	g.EditView.Render(w, vd)
}

// Update ,,,
// GET /galleries/:id/edit
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	vd.Yield = gallery

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery Not Found", http.StatusNotFound)
		return
	}

	var galleryForm GalleryForm
	if err := parseForm(r, &galleryForm); err != nil {
		log.Println("galleries.Udpate() parseForm ERROR:", err)
		vd.SetAlert(err)
		g.EditView.Render(w, vd)
		return
	}

	gallery.Title = galleryForm.Title
	if err := g.gs.Update(gallery); err != nil {
		log.Println("g.gs.Update() ERROR:", err)
		vd.SetAlert(err)
		g.EditView.Render(w, vd)
		return
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Gallery Successfully Updated!",
	}

	g.EditView.Render(w, vd)
}
