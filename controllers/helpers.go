package controllers

import (
	"net/http"

	schema "github.com/gorilla/Schema"
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(dst, r.PostForm)
	if err != nil {
		return err
	}

	return nil
}
