package main

import (
    "errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var templates = template.Must(template.ParseFiles(loadTemplates()...))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
    name := p.Title + ".txt"
    return os.WriteFile(name, p.Body, 0600)
}

func loadTemplates() []string {
    viewsDir := "views"
    files := make([]string, 0)
    entries, err := os.ReadDir(viewsDir)
    if err != nil {
        fmt.Printf("err: %s", err.Error())
        return files
    }
    for _, e := range entries {
        if strings.Contains(e.Name(), ".html") {
            name := viewsDir + "/" + e.Name()
            files = append(files, name)
        }
    }
    fmt.Println(files)
    return files
}

func loadPage(title string) (*Page, error) {
    name := title + ".txt"
    body, err := os.ReadFile(name)
    if err != nil {
        return nil, err
    }
    return &Page{title, body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    if err := templates.ExecuteTemplate(w, tmpl+".html", p); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("invalid page title")
    }
    return m[2], nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    if err := p.save(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        title, err := getTitle(w, r)
        if err != nil {
            fmt.Printf("err: %s", err.Error())
            return
        }
        fn(w, r, title)
    }
}

func main() {
    logger := log.New(os.Stdout, "wiki: ", log.LstdFlags)

    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))

    logger.Printf("server listening on: http://localhost:8080\n")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
