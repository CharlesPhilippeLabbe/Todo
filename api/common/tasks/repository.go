package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"todo/common/storage"
)

type Repository struct {
	table string
	db    *storage.Database
}

type ListFunc func(*Task)

func NewRepository(db *storage.Database) (*Repository, error) {
	err := db.CreateTable(&storage.Query{
		Table: "tasks",
		Columns: []string{
			"id varchar(32) not null PRIMARY KEY",
			"list varchar(32)",
			"category varchar(32)",
			"name text not null",
			"priority varchar(32)",
		}})

	if err != nil {
		return nil, fmt.Errorf("could not create table: %w", err)
	}

	return &Repository{"tasks", db}, nil
}

func (r *Repository) AddTask(ctx context.Context, list, category, name string) (string, error) {
	id, err := r.db.CreateId()
	if err != nil {
		return "", err
	}

	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO tasks (id, list, category, name) VALUES (?, ?, ?, ?);",
		id, list, category, name)

	return id, err
}

func (r *Repository) AllLists(ctx context.Context) ([]string, error) {
	res, err := r.db.QueryContext(ctx,
		"SELECT list from tasks group by list")

	if err != nil {
		return nil, err
	}
	defer res.Close()
	if res == nil {
		return nil, fmt.Errorf("result is nil")
	}

	l := make([]string, 0)
	for res.Next() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var s string
		err = res.Scan(&s)
		if err != nil {
			return nil, err
		}
		l = append(l, s)
	}

	return l, nil
}

func (r *Repository) ListCategory(ctx context.Context, list, category string) ([]*Task, error) {

	res, err := r.db.QueryContext(ctx,
		"SELECT id, list, category, name from tasks where list = ? and category = ? order by id desc",
		list, category)

	if err != nil {
		return nil, err
	}
	defer res.Close()

	return createList(ctx, res)
}

func (r *Repository) List(ctx context.Context, list string, f ListFunc) error {

	res, err := r.db.QueryContext(ctx,
		"SELECT id, list, category, name from tasks where list = ? order by id desc",
		list)

	if err != nil {
		return err
	}
	defer res.Close()

	return tasksFromResult(ctx, res, f)
}

func (r *Repository) Put(ctx context.Context, id, newCategory string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tasks set category = ? where id = ?",
		newCategory, id)

	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE from tasks where id = ?",
		id)

	return err
}

func (r *Repository) Get(ctx context.Context, list, id string) (*Task, error) {
	res, err := r.db.QueryContext(ctx,
		"SELECT id, list, category, name from tasks where list = ? and id = ?",
		list, id)

	if err != nil {
		return nil, err
	}
	defer res.Close()

	l, err := createList(ctx, res)
	if err != nil {
		return nil, err
	} else if len(l) == 0 {
		return nil, nil
	}

	return l[0], nil
}

func createList(ctx context.Context, res *sql.Rows) ([]*Task, error) {
	tasks := make([]*Task, 0)

	err := tasksFromResult(ctx, res, func(t *Task) {
		tasks = append(tasks, t)
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func tasksFromResult(ctx context.Context, res *sql.Rows, f ListFunc) error {
	if res == nil {
		return fmt.Errorf("result is nil")
	}

	var err error
	for res.Next() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		t := Task{}
		err = res.Scan(&t.Id, &t.List, &t.Category, &t.Name)
		if err != nil {
			return err
		}
		f(&t)
	}

	return nil
}
