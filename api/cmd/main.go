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
	List   string
	Values map[string]string
	Errors map[string]string
}

func newFormData(list string) FormData {
	return FormData{
		List:   list,
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
		Form: newFormData("default"),
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

	mux.HandleFunc("GET /api/{list}", func(w http.ResponseWriter, r *http.Request) {
		list := r.PathValue("list")
		if list == "" {
			list = "default"
		}
		l, err := tc.ListTasks(r.Context(), list)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		render(w, "index", &Page{
			Form: newFormData(list),
			Data: l,
		})
	})

	mux.HandleFunc("POST /api/{list}/tasks", func(w http.ResponseWriter, r *http.Request) {
		list := r.PathValue("list")
		name := r.FormValue("name")
		//
		t, err := tc.NewTask(r.Context(), list, "ToDo", name)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			formData := newFormData(list)
			formData.Values["name"] = name
			formData.Errors["task"] = err.Error()
			render(w, "form", formData)
			return
		}

		log.Printf("New Task: %v\n", t)
		render(w, "form", newFormData(list))
		render(w, "oob-task", t)
	})

	mux.HandleFunc("PUT /api/{list}/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		list := r.PathValue("list")
		id := r.PathValue("id")
		move := r.FormValue("move")
		l, err := tc.MoveTask(r.Context(), list, id, move)
		if err != nil {
			log.Println(err)
			return
		}
		render(w, "display", l)
	})

	mux.HandleFunc("POST /api/lists", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		//
		//t, err := tc.NewTask(r.Context(), name, "ToDo", name)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			formData := newFormData(name)
			formData.Values["name"] = name
			formData.Errors["task"] = err.Error()
			render(w, "form", formData)
			return
		}

		w.Header().Add("Hx-Trigger-After-Swap", fmt.Sprintf(`{"afterAdd":"%s"}`, name))
		render(w, "addList", newFormData(""))
		l, err := tc.ListTasks(r.Context(), name)
		if err != nil {
			log.Println(err)
			//TODO
			return
		}
		p := &Page{
			Data: l,
			Form: newFormData(name),
		}

		render(w, "display-oob", p)
		//render(w, "oob-task", t)
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
