package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/arnoldokoth/lenslocked.com/context"
	"github.com/gorilla/csrf"
)

var (
	templateDir string = "views/"
	layoutDir   string = "views/layouts/"
	templateExt string = ".gohtml"
)

// NewView ...
func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)
	t, err := template.New("").Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", errors.New("csrfField Not Implemented")
			},
		}).ParseFiles(files...)
	if err != nil {
		log.Fatalln("views.NewView() ERROR:", err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

// View ...
type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

// Render ...
func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		// do nothing
		vd = d
	default:
		// move whatever data was provided e.g. a string to the Yield field in
		// the Data struct
		vd = Data{
			Yield: data,
		}
	}

	if alert := getAlert(r); alert != nil {
		vd.Alert = alert
		clearAlert(w)
	}

	vd.User = context.User(r.Context())

	var buffer bytes.Buffer

	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})

	if err := tpl.ExecuteTemplate(&buffer, v.Layout, vd); err != nil {
		log.Println("view.Render() ERROR:", err)
		http.Error(w, "Something Went Wrong. If the problem persists, please email support@lenslocked.com", http.StatusInternalServerError)
		return
	}

	io.Copy(w, &buffer)
}

// layoutFiles returns the list of files in
// the views/layouts directory
func layoutFiles() []string {
	files, err := filepath.Glob(layoutDir + "*" + templateExt)
	if err != nil {
		log.Fatalln("layoutFiles() ERROR:", err)
	}

	return files
}

// addTemplatePath takes in a slice of strings
// representing file paths for templates and
// prepends the templateDir to each string
func addTemplatePath(files []string) {
	for idx, file := range files {
		files[idx] = templateDir + file
	}
}

func addTemplateExt(files []string) {
	for idx, file := range files {
		files[idx] = file + templateExt
	}
}
