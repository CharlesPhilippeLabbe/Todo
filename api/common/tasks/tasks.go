package tasks

import (
	"fmt"
	"slices"
)

var (
	ErrTaskDoesNotExist error = fmt.Errorf("task not found")
)

type Task struct {
	Name string
}

func NewTask(name string) Task {
	return Task{
		Name: name,
	}
}

type Tasks struct {
	Name  string
	Tasks []Task
}

func (d *Tasks) DeleteTask(name string) bool {
	i := d.HasTask(name)
	if i < 0 {
		return false
	}
	d.Tasks = slices.Delete(d.Tasks, i, i+1)
	return true
}
func (d *Tasks) HasTask(name string) int {
	for i, task := range d.Tasks {
		if task.Name == name {
			return i
		}
	}

	return -1
}

type Data struct {
	Todo  Tasks
	Doing Tasks
	Done  Tasks
}

func NewData() Data {
	return Data{
		Todo: Tasks{
			Name: "ToDo",
			Tasks: []Task{
				NewTask("test"),
				NewTask("John"),
				NewTask("Claire"),
			},
		},
		Doing: Tasks{Name: "Doing"},
		Done:  Tasks{Name: "Done"},
	}
}

func (d *Data) MoveTask(name, direction string) error {
	var target *Tasks
	if d.Todo.DeleteTask(name) {
		if direction == "forward" {
			target = &d.Doing
		}
	} else if d.Doing.DeleteTask(name) {
		if direction == "forward" {
			target = &d.Done
		} else {
			target = &d.Todo
		}
	} else if d.Done.DeleteTask(name) {
		if direction == "back" {
			target = &d.Doing
		}
	} else {
		return ErrTaskDoesNotExist
	}
	if target != nil && target.HasTask(name) < 0 {
		target.Tasks = append(target.Tasks, NewTask(name))
	}

	return nil
}
