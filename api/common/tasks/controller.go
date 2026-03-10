package tasks

import (
	"context"
	"errors"
	"fmt"
)

type Controller struct {
	r *Repository
}

func NewController(r *Repository) *Controller {
	return &Controller{r}
}

func (c *Controller) NewTask(ctx context.Context, list, category, name string) (*Task, error) {
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

func (c *Controller) MoveTask(ctx context.Context, list, id, direction string) (*List, error) {
	//TODO optimize...
	l, err := c.ListTasks(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("could not list tasks: %w", err)
	}
	newCategory, err := l.MoveTask(id, direction)
	if errors.Is(err, ErrTargetDoesNotExist) {
		err = c.r.Delete(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("could not delete task: %w", err)
		}

	} else if err != nil {
		return nil, err
	}

	err = c.r.Put(ctx, id, newCategory)
	if err != nil {
		return nil, fmt.Errorf("could not update task: %w", err)
	}
	return l, nil
}
