package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
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
	Data tasks.Data
	Form FormData
}

func newPage() Page {

	return Page{
		Data: tasks.NewData(),
		Form: newFormData(),
	}
}

func main() {

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
		render(w, "index", page)
	})

	mux.HandleFunc("POST /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		if page.Data.Todo.HasTask(name) >= 0 {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Errors["task"] = "Already exists"

			w.WriteHeader(422)
			render(w, "form", formData)
			return
		}
		task := tasks.NewTask(name)
		page.Data.Todo.Tasks = append(page.Data.Todo.Tasks, task)

		render(w, "form", newFormData())
		render(w, "oob-task", task)
	})
	mux.HandleFunc("PUT /api/tasks/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		move := r.FormValue("move")

		err := page.Data.MoveTask(name, move)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("not found"))
			if err != nil {
				log.Println(err)
			}
		}
		render(w, "display", page.Data)
	})

	mux.HandleFunc("DELETE /data", func(w http.ResponseWriter, r *http.Request) {
		page = newPage()
		render(w, "index", page)
	})

	err = http.ListenAndServe(":8080", LoggingMiddleware(mux))
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
