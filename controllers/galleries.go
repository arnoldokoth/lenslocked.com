package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/arnoldokoth/lenslocked.com/context"
	"github.com/arnoldokoth/lenslocked.com/models"
	"github.com/arnoldokoth/lenslocked.com/views"
	"github.com/gorilla/mux"
)

const (
	showGallery     = "show_gallery"
	editGallery     = "edit_gallery"
	maxMultipartMem = 1 << 20
)

// NewGalleries ...
func NewGalleries(gs models.GalleryService, router *mux.Router) *Galleries {
	return &Galleries{
		IndexView:  views.NewView("bootstrap", "galleries/index"),
		CreateView: views.NewView("bootstrap", "galleries/new"),
		ShowView:   views.NewView("bootstrap", "galleries/show"),
		EditView:   views.NewView("bootstrap", "galleries/edit"),
		gs:         gs,
		router:     router,
	}
}

// Galleries ...
type Galleries struct {
	IndexView  *views.View
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
	g.CreateView.Render(w, r, nil)
}

// Index ...
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		http.Error(w, ErrGeneric.Error(), http.StatusInternalServerError)
		return
	}
	vd.Yield = galleries
	// fmt.Fprintln(w, galleries)
	g.IndexView.Render(w, r, vd)
}

// Create ...
// POST /galleries/new
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var galleryForm GalleryForm
	if err := parseForm(r, &galleryForm); err != nil {
		log.Println("galleries.Create() ERROR:", err)
		vd.SetAlert(err)
		g.CreateView.Render(w, r, vd)
		return
	}

	user := context.User(r.Context())

	gallery := models.Gallery{
		Title:  galleryForm.Title,
		UserID: user.ID,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.CreateView.Render(w, r, vd)
		return
	}

	url, err := g.router.Get(editGallery).URL("id", fmt.Sprintf("%+v", gallery.ID))
	if err != nil {
		// TODO: Redirect To Galleries Index Page
		http.Redirect(w, r, "/galleries", http.StatusFound)
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

	g.ShowView.Render(w, r, vd)
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
	g.EditView.Render(w, r, vd)
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
		g.EditView.Render(w, r, vd)
		return
	}

	gallery.Title = galleryForm.Title
	if err := g.gs.Update(gallery); err != nil {
		log.Println("g.gs.Update() ERROR:", err)
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Gallery Successfully Updated!",
	}

	g.EditView.Render(w, r, vd)
}

// Upload ...
// POST /galleries/:id/images
func (g *Galleries) Upload(w http.ResponseWriter, r *http.Request) {
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

	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

		defer file.Close()

		galleryPath := fmt.Sprintf("images/galleries/%v/", gallery.ID)
		err = os.MkdirAll(galleryPath, 0755)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

		dst, err := os.Create(galleryPath + f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}

		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}

	fmt.Fprintln(w, "Files Successfully Uploaded")
}

// Delete ,,,
// POST /galleries/:id/delete
func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
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

	err = g.gs.Delete(gallery.ID)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	// TODO: Redirect To Galleries Index Page
	http.Redirect(w, r, "/galleries", http.StatusFound)
}
