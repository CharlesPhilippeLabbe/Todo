package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
	"todo/common/storage"
	"todo/common/tasks"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any) error {

	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {

	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data *tasks.List
	Form FormData
}

func newPage() Page {

	return Page{
		Data: tasks.NewList("default"),
		Form: newFormData(),
	}
}

var port int

func main() {
	port = *flag.Int("port", 8080, "Port number")

	db, err := storage.NewSqlite("tmp/test.db")
	if err != nil {
		log.Fatalf("could not create database: %v", err)
		return
	}

	repo, err := tasks.NewRepository(db)
	if err != nil {
		log.Fatalf("could not create repository: %v", err)
		return
	}
	tc := tasks.NewController(repo)

	page := newPage()
	templates := newTemplate()

	target, err := url.Parse("http://localhost:1313")
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	mux := http.NewServeMux()

	// Registering a handler for a specific method and path with a variable
	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	render := func(w http.ResponseWriter, t string, d any) {
		err := templates.Render(w, t, d)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
		l, err := tc.ListTasks(r.Context(), "default")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		render(w, "index", &Page{
			Form: newFormData(),
			Data: l,
		})
	})

	mux.HandleFunc("POST /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")

		//
		t, err := tc.NewTask(r.Context(), "default", "ToDo", name)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			formData := newFormData()
			formData.Values["name"] = name
			formData.Errors["task"] = err.Error()
			render(w, "form", formData)
			return
		}

		log.Printf("New Task: %v\n", t)
		render(w, "form", newFormData())
		render(w, "oob-task", t)
	})

	mux.HandleFunc("PUT /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		move := r.FormValue("move")
		l, err := tc.MoveTask(r.Context(), "default", id, move)
		if err != nil {
			log.Println(err)
			return
		}
		render(w, "display", l)
	})

	mux.HandleFunc("DELETE /data", func(w http.ResponseWriter, r *http.Request) {
		page = newPage()
		render(w, "index", page)
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), LoggingMiddleware(mux))
	if err != nil {
		log.Println(err)
	}
}
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s took %v\n", r.Method, r.URL.Path, r.Proto, time.Since(start))
	})
}
