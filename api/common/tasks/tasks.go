package tasks

import (
	"fmt"
	"slices"
)

var (
	ErrTaskDoesNotExist     error = fmt.Errorf("task not found")
	ErrTargetDoesNotExist   error = fmt.Errorf("target does not exit")
	ErrUnsupportedDirection error = fmt.Errorf("unsupported direction")
)

type Task struct {
	Id       string  `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	List     string  `json:"list,omitempty"`
	Category string  `json:"category,omitempty"`
	Priority *string `json:"priority,omitempty"`
}

type Position string

const (
	Above Position = "above"
	Below Position = "below"
)

type TargetTask struct {
	Task
	Target   *Task
	Position Position
}

func NewTask(name string) Task {
	return Task{
		Name: name,
	}
}

type Tasks struct {
	Name  string
	List  string
	Tasks []Task
}

func NewTasks(list, name string) *Tasks {
	return &Tasks{
		List:  list,
		Name:  name,
		Tasks: make([]Task, 0),
	}
}

func (d *Tasks) DeleteTask(id string) *Task {
	if d == nil {
		return nil
	}
	i := d.HasTask(id)
	if i < 0 {
		return nil
	}
	t := d.Tasks[i]
	d.Tasks = slices.Delete(d.Tasks, i, i+1)
	return &t
}

func (d *Tasks) HasTask(id string) int {
	if d == nil {
		return -1
	}

	for i, task := range d.Tasks {
		if task.Id == id {
			return i
		}
	}

	return -1
}

func (d *Tasks) AddTask(t *Task) {

	if t == nil {
		return
	}

	d.Tasks = append(d.Tasks, *t)
}

type List struct {
	Name  string            `json:"name,omitempty"`
	Tasks map[string]*Tasks `json:"tasks,omitempty"`
}

func NewList(name string) *List {

	return &List{
		Name: name,
		Tasks: map[string]*Tasks{
			"ToDo":  NewTasks(name, "ToDo"),
			"Doing": NewTasks(name, "Doing"),
			"Done":  NewTasks(name, "Done"),
		},
	}
}

func (l *List) AddTask(t *Task) {
	if t == nil {
		return
	}
	tasks, ok := l.Tasks[t.Category]
	if !ok {
		tasks = NewTasks(t.List, t.Category)
		l.Tasks[t.Category] = tasks
	}

	tasks.AddTask(t)
}

func (l *List) MoveTask(id, direction string) (*Task, error) {
	var target *Tasks
	todo := l.Tasks["ToDo"]
	doing := l.Tasks["Doing"]
	done := l.Tasks["Done"]

	t := todo.DeleteTask(id)

	if t != nil {
		if direction == "forward" {
			target = doing
		}
	} else if t = doing.DeleteTask(id); t != nil {
		if direction == "forward" {
			target = done
		} else {
			target = todo
		}
	} else if t = done.DeleteTask(id); t != nil {
		if direction == "back" {
			target = doing
		}
	} else {
		return nil, ErrTaskDoesNotExist
	}
	if target == nil {
		return nil, ErrTargetDoesNotExist
	}

	if target.HasTask(id) < 0 {
		t.Category = target.Name
		target.Tasks = append(target.Tasks, *t)
	}

	return t, nil
}
