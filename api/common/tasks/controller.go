package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Controller struct {
	r *Repository
}

func NewController(r *Repository) *Controller {
	return &Controller{r}
}

func (c *Controller) NewTask(ctx context.Context, list, category, name string) (*Task, error) {
	name = strings.TrimSpace(name)
	list = strings.TrimSpace(list)
	category = strings.TrimSpace(category)

	id, err := c.r.AddTask(ctx, list, category, name)
	if err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return &Task{
		Id:       id,
		List:     list,
		Category: category,
		Name:     name,
	}, nil
}

func (c *Controller) AllLists(ctx context.Context) ([]string, error){
	return c.r.AllLists(ctx)
}

func (c *Controller) ListCategory(ctx context.Context, list, category string) ([]*Task, error) {
	tasks, err := c.r.ListCategory(ctx, list, category)
	if err != nil {
		return nil, fmt.Errorf("could not list tasks: %w", err)
	}
	return tasks, nil
}

func (c *Controller) ListTasks(ctx context.Context, name string) (*List, error) {

	list := NewList(name)

	err := c.r.List(ctx, name, list.AddTask)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (c *Controller) GetTask(ctx context.Context, list, id string) (*Task, error) {
	t, err := c.r.Get(ctx, list, id)
	if err != nil {
		return nil, err
	} else if t == nil {
		return nil, fmt.Errorf("task %s not found: %w", id, ErrTaskDoesNotExist)
	}

	return t, nil
}

func (c *Controller) MoveTask(ctx context.Context, list, id, direction string) (*Task, error) {
	t, err := c.GetTask(ctx, list, id)
	if err != nil {
		return t, nil
	}
	//TODO remove the hardcoded stuff
	action := t.Category + "_" + direction
	switch action {
	case "ToDo_forward":
		t.Category = "Doing"
	case "Doing_back":
		t.Category = "ToDo"
	case "Doing_forward":
		t.Category = "Done"
	case "Done_back":
		t.Category = "Doing"
	case "ToDo_back":
		fallthrough
	case "Done_forward":
		err = ErrTargetDoesNotExist
	}
	//TODO optimize...
	if errors.Is(err, ErrTargetDoesNotExist) {
		err = c.r.Delete(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("could not delete task: %w", err)
		}
		return nil, nil
	}

	err = c.r.Put(ctx, id, t.Category)
	if err != nil {
		return nil, fmt.Errorf("could not update task: %w", err)
	}
	return t, nil
}
