package main

import (
	"fmt"
	"html/template"
	"log"
	// "mime"
	"net/http"
	"os/exec"
)

// The server uses the external command: wkhtmltopdf - -

var doc = `<html><p>Hello, {{ .Name }}!</p></html>`

var tmpl = template.Must(template.New("doc").Parse(doc))

type Attrs struct {
	Name string
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:] // Use the path as a name, removing the initial /
	if path == "" {
		path = "Golang PDF" // Sane default
	}

	// NOTE: to force a download of the generated PDF, add the following
	// lines and import the "mime" package
	// h := fmt.Sprintf("attachment; filename=%s", "output.pdf")
	// w.Header().Set("Content-Disposition", h)
	// w.Header().Set("Content-Type", mime.TypeByExtension(".pdf"))

	// Prepare the command to read from std in and write to std out
	wkhtmltopdf := exec.Command("wkhtmltopdf", "-", "-")
	wkhtmltopdf.Stdout = w

	input, err := wkhtmltopdf.StdinPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Begin the command
	if err = wkhtmltopdf.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the html template to std in and close
	if err = tmpl.Execute(input, Attrs{Name: path}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = input.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Wait for the command to end
	if err = wkhtmltopdf.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	port := 8081
	log.Printf("Starting the server on port %d", port)
	http.HandleFunc("/", Handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
