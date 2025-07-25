package validation

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNameRequired   = errors.New("name is required")
	ErrNameTooLong    = errors.New("name must be 100 characters or less")
	ErrEmailRequired  = errors.New("email is required")
	ErrEmailInvalid   = errors.New("email format is invalid")
	ErrAgeInvalid     = errors.New("age must be between 1 and 149")
	ErrUserIDRequired = errors.New("user ID is required")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateCreateUser validates user creation data
func ValidateCreateUser(name, email string, age int32) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validateAge(age); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateUser validates user update data
func ValidateUpdateUser(id, name, email string, age int32) error {
	if err := validateUserID(id); err != nil {
		return err
	}

	// For updates, only validate provided fields
	if name != "" {
		if err := validateName(name); err != nil {
			return err
		}
	}
	if email != "" {
		if err := validateEmail(email); err != nil {
			return err
		}
	}
	if age > 0 {
		if err := validateAge(age); err != nil {
			return err
		}
	}
	return nil
}

// ValidateUserID validates user ID
func ValidateUserID(id string) error {
	return validateUserID(id)
}

func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	if len(name) > 100 {
		return ErrNameTooLong
	}
	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}
	return nil
}

func validateAge(age int32) error {
	if age <= 0 || age >= 150 {
		return ErrAgeInvalid
	}
	return nil
}

func validateUserID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrUserIDRequired
	}
	return nil
}
