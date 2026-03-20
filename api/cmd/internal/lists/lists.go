package lists

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"
	"todo/common/tasks"
)

var templates *Templates = newTemplate()

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

type Controller struct {
	templates *Templates
	tc        *tasks.Controller
}

func NewController(tc *tasks.Controller) *Controller {
	return &Controller{
		templates: newTemplate(),
		tc:        tc,
	}
}

func (c *Controller) render(w http.ResponseWriter, t string, d any) {
	err := templates.Render(w, t, d)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {

	pageList, err := c.tc.AllLists(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.render(w, "index", pageList)
}

func (c *Controller) Selection(w http.ResponseWriter, r *http.Request) {
	l, err := c.tc.ListTasks(r.Context(), "default")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	c.render(w, "list-oob", &Page{
		Data: l,
		Form: newFormData("default"),
	})

	lists := r.URL.Query()["lists"]

	for _, l := range lists {
		log.Println(l)
		tl, err := c.tc.ListTasks(r.Context(), l)
		if err != nil {
			log.Printf("Could not retreive list %s: %v\n", l, err)
			continue
		}

		c.render(w, "list-oob", &Page{
			Data: tl,
			Form: newFormData(l),
		})
	}
	//c.render(w, "listSelection", pageList)
}

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	list := r.PathValue("list")
	if list == "" {
		list = "default"
	}
	l, err := c.tc.ListTasks(r.Context(), list)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	c.render(w, "list-oob", &Page{
		Form: newFormData(list),
		Data: l,
	})
}

func (c *Controller) RemoveList(w http.ResponseWriter, r *http.Request) {

	list := r.PathValue("list")
	w.Header().Add("Hx-Trigger-After-Swap", fmt.Sprintf(`{"afterRemove":"%s"}`, list))
	c.render(w, "deleteList", list)
}

func (c *Controller) AddTask(w http.ResponseWriter, r *http.Request) {
	list := r.PathValue("list")
	name := r.FormValue("name")
	//
	t, err := c.tc.NewTask(r.Context(), list, "ToDo", name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		formData := newFormData(list)
		formData.Values["name"] = name
		formData.Errors["task"] = err.Error()
		c.render(w, "form", formData)
		return
	}

	log.Printf("New Task: %v\n", t)
	c.render(w, "form", newFormData(list))
	c.render(w, "oob-task", t)
}

func (c *Controller) MoveTask(w http.ResponseWriter, r *http.Request) {
	list := r.PathValue("list")
	id := r.PathValue("id")
	move := r.FormValue("move")
	t, err := c.tc.MoveTask(r.Context(), list, id, move)
	if err != nil {
		log.Println(err)
		return
	} else if t == nil {
		return
	}

	c.render(w, "oob-task", t)
}

func (c *Controller) AddList(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		name = r.PathValue("list")
	}
	name = strings.TrimSpace(name)
	//
	w.Header().Add("Hx-Trigger-After-Swap", fmt.Sprintf(`{"afterAdd":"%s"}`, name))
	c.render(w, "addList", newFormData(""))
	l, err := c.tc.ListTasks(r.Context(), name)
	if err != nil {
		log.Println(err)
		//TODO
		return
	}
	p := &Page{
		Data: l,
		Form: newFormData(name),
	}

	c.render(w, "list-oob", p)
	//c.render(w, "selectListInput-oob", name)
	//c.render(w, "oob-task", t)
}
func (c *Controller) ToggleList(w http.ResponseWriter, r *http.Request) {
	list := r.PathValue("list")
	on := r.FormValue(list)

	if on == "on" {
		c.AddList(w, r)
	} else {
		c.RemoveList(w, r)
	}

}
