package validation

import (
	"errors"
	"regexp"
	"strings"
)

type SignupInput struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

func (s *SignupInput) Validate() error {
	s.Username = strings.TrimSpace(s.Username)
	s.Password = strings.TrimSpace(s.Password)

	if s.Username == "" {
		return errors.New("username is required")
	}
	if len(s.Username) < 3 || len(s.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(s.Username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	if s.Password == "" {
		return errors.New("password is required")
	}
	if len(s.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	return nil
}
