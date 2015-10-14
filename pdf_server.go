package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// Target command: wkhtmltopdf - -

var doc = `<html><p>Hello, {{ .Name }}!</p></html>`

var tmpl = template.Must(template.New("doc").Parse(doc))

type Attrs struct {
	Name string
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Get the name from the path
	path := r.URL.Path[1:] // Remove initial /
	if path == "" {
		// Sane default
		path = "Golang PDF"
	}

	attrs := Attrs{Name: path}

	// Prepare the command to read from std in and write to std out
	wkhtmltopdf := exec.Command("wkhtmltopdf", "-", "-")

	input, err := wkhtmltopdf.StdinPipe()
	if err != nil {
		log.Panic("in err", err)
	}

	output, err := wkhtmltopdf.StdoutPipe()
	if err != nil {
		log.Panic("out err", err)
	}

	// Begin the command
	if err = wkhtmltopdf.Start(); err != nil {
		log.Panic("run err", err)
	}

	// Write the html template to std in and close
	if err = tmpl.Execute(input, attrs); err != nil {
		log.Panic("template err", err)
	}

	if err = input.Close(); err != nil {
		log.Panic("close err", err)
	}

	// Read the generated PDF from std out
	b, err := ioutil.ReadAll(output)
	if err != nil {
		log.Fatal("io err", err)
	}

	// End the command
	if err = wkhtmltopdf.Wait(); err != nil {
		log.Fatal("wait err", err)
	}

	// Convert to a read seeker
	// TODO Can't the stdout be converted to a readseeker without reading?
	f := bytes.NewReader(b)

	filename := "output.pdf"
	h := fmt.Sprintf("attachment; filename=%s", filename)
	w.Header().Set("Content-Disposition", h)

	// TODO Set mimetype

	http.ServeContent(w, r, filename, time.Now(), f)
}

func main() {
	port := 8081
	log.Printf("Starting the server on port %d", port)
	http.HandleFunc("/", Handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
