package error

import (
	"net/http"
	"text/template"
)

type Error struct {
	Status int
	Error  string
}

func RenderError(w http.ResponseWriter, r *http.Request, status int, error string) error {
	tmpl, err := template.ParseFiles("templates/error.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, Error{
		Status: status,
		Error:  error,
	})

	return nil
}
