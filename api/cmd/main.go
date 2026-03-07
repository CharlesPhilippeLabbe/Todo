package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http/httputil"
	"net/url"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

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
	for i, contact := range d.Tasks {
		if contact.Name == name {
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

	e := echo.New()
	e.Use(middleware.Logger())

	page := newPage()
	e.Renderer = newTemplate()

	target, err := url.Parse("http://localhost:1313")
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	e.Any("/*", func(c echo.Context) error {
		req := c.Request()

		proxy.ServeHTTP(c.Response().Writer, req)
		return nil
	})

	e.GET("/api", func(c echo.Context) error {
		err := c.Render(200, "index", page)
		if err != nil {
			fmt.Println(err)
		}
		return err
	})

	e.POST("/api/contacts", func(c echo.Context) error {

		name := c.FormValue("name")
		if page.Data.Todo.hasTask(name) >= 0 {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Errors["email"] = "Already exists"

			return c.Render(422, "form", formData)
		}
		contact := newTask(name)
		page.Data.Todo.Tasks = append(page.Data.Todo.Tasks, contact)

		c.Render(200, "form", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.PUT("/api/task/:name", func(c echo.Context) error {
		name := c.Param("name")
		if page.Data.Todo.deleteTask(name) {
			if page.Data.Doing.hasTask(name) < 0 {
				page.Data.Doing.Tasks = append(page.Data.Doing.Tasks, newTask(name))
			}
		} else if page.Data.Doing.deleteTask(name) {
			if page.Data.Done.hasTask(name) < 0 {
				page.Data.Done.Tasks = append(page.Data.Done.Tasks, newTask(name))
			}
		}else if !page.Data.Done.deleteTask(name) {
			return c.HTML(404, "not found")
		}
		return c.Render(200, "display", page.Data)
	})

	e.DELETE("/data", func(c echo.Context) error {
		page = newPage()
		err := c.Render(200, "index", page)
		if err != nil {
			fmt.Println(err)
		}
		return err
	})

	e.Logger.Fatal(e.Start(":6969"))
}
