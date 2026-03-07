package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"time"
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

type Task struct {
	Name string
}

func newTask(name string) Task {
	return Task{
		Name: name,
	}
}

type Tasks struct {
	Name  string
	Tasks []Task
}

type Data struct {
	Todo  Tasks
	Doing Tasks
	Done  Tasks
}

func newData() Data {
	return Data{
		Todo: Tasks{
			Name: "ToDo",
			Tasks: []Task{
				newTask("test"),
				newTask("John"),
				newTask("Claire"),
			},
		},
		Doing: Tasks{Name: "Doing"},
		Done:  Tasks{Name: "Done"},
	}
}

func (d *Tasks) deleteTask(name string) bool {
	i := d.hasTask(name)
	if i < 0 {
		return false
	}
	d.Tasks = slices.Delete(d.Tasks, i, i+1)
	return true
}
func (d *Tasks) hasTask(name string) int {
	for i, task := range d.Tasks {
		if task.Name == name {
			return i
		}
	}

	return -1
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
	Data Data
	Form FormData
}

func newPage() Page {

	return Page{
		Data: newData(),
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
		if page.Data.Todo.hasTask(name) >= 0 {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Errors["task"] = "Already exists"

			w.WriteHeader(422)
			render(w, "form", formData)
			return
		}
		task := newTask(name)
		page.Data.Todo.Tasks = append(page.Data.Todo.Tasks, task)

		render(w, "form", newFormData())
		render(w, "oob-task", task)
	})
	mux.HandleFunc("PUT /api/tasks/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		move := r.FormValue("move")

		var target *Tasks
		if page.Data.Todo.deleteTask(name) {
			if move == "forward" {
				target = &page.Data.Doing
			}
		} else if page.Data.Doing.deleteTask(name) {
			if move == "forward" {
				target = &page.Data.Done
			} else {
				target = &page.Data.Todo
			}
		} else if page.Data.Done.deleteTask(name) {
			if move == "back" {
				target = &page.Data.Doing
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("not found"))
			if err != nil {
				log.Println(err)
			}
			return
		}
		if target != nil && target.hasTask(name) < 0 {
			target.Tasks = append(target.Tasks, newTask(name))
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
