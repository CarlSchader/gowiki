package main

import (
	"html/template"
	"log"
	"io/ioutil"
	"net/http"
	"regexp"
	"errors"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func checkError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil
}

func renderTemplate(w http.ResponseWriter, page *Page, templatePath string) {
	err := templates.ExecuteTemplate(w, templatePath, page)
	checkError(w, err)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, err := getTitle(w, r)
		checkError(w, err)
		fn(w, r, title)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {	
	page, err := loadPage(title)
	
	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
	}
	
	renderTemplate(w, page, "view.html")
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	
	if err != nil {
		page = &Page{Title: title, Body: []byte("New page")}
	}

	renderTemplate(w, page, "edit.html")
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	page := &Page{Title: title, Body: []byte(r.FormValue("body"))}
	err := page.save()
	checkError(w, err)
	
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}