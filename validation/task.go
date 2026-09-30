package validation

import (
	"errors"
	"strings"
)

type TaskInput struct {
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}

func (t *TaskInput) Validate() error {
	t.Title = strings.TrimSpace(t.Title)
	t.Description = strings.TrimSpace(t.Description)

	if t.Title == "" {
		return errors.New("title is required")
	}
	if len(t.Title) < 3 || len(t.Title) > 150 {
		return errors.New("title must be between 3 and 150 characters")
	}
	if len(t.Description) > 500 {
		return errors.New("description must not exceed 500 characters")
	}
	return nil
}
