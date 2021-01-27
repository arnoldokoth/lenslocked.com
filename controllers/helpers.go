package controllers

import (
	"net/http"
	"net/url"

	schema "github.com/gorilla/Schema"
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	return parseValues(r.PostForm, dst)
}

func parseURLParams(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	return parseValues(r.Form, dst)
}

func parseValues(values url.Values, dst interface{}) error {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(dst, values)
	if err != nil {
		return err
	}

	return nil
}
