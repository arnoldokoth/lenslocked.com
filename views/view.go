package views

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var (
	templateDir        = "views/"
	layoutDir   string = "views/layouts/"
	templateExt string = ".gohtml"
)

// NewView ...
func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
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
	v.Render(w, nil)
}

// Render ...
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
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
