package validation

import (
	"errors"
	"log/slog"
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
	slog.Debug("Validating create user request", "name", name, "email", email, "age", age)

	if err := validateName(name); err != nil {
		slog.Warn("Name validation failed", "name", name, "error", err)
		return err
	}
	if err := validateEmail(email); err != nil {
		slog.Warn("Email validation failed", "email", email, "error", err)
		return err
	}
	if err := validateAge(age); err != nil {
		slog.Warn("Age validation failed", "age", age, "error", err)
		return err
	}

	slog.Debug("Create user validation passed", "name", name, "email", email, "age", age)
	return nil
}

// ValidateUpdateUser validates user update data
func ValidateUpdateUser(id, name, email string, age int32) error {
	slog.Debug("Validating update user request", "user_id", id, "name", name, "email", email, "age", age)

	if err := validateUserID(id); err != nil {
		slog.Warn("User ID validation failed", "user_id", id, "error", err)
		return err
	}

	// For updates, only validate provided fields
	if name != "" {
		if err := validateName(name); err != nil {
			slog.Warn("Name validation failed in update", "name", name, "error", err)
			return err
		}
	}
	if email != "" {
		if err := validateEmail(email); err != nil {
			slog.Warn("Email validation failed in update", "email", email, "error", err)
			return err
		}
	}
	if age > 0 {
		if err := validateAge(age); err != nil {
			slog.Warn("Age validation failed in update", "age", age, "error", err)
			return err
		}
	}

	slog.Debug("Update user validation passed", "user_id", id, "name", name, "email", email, "age", age)
	return nil
}

// ValidateUserID validates user ID
func ValidateUserID(id string) error {
	slog.Debug("Validating user ID", "user_id", id)

	err := validateUserID(id)
	if err != nil {
		slog.Warn("User ID validation failed", "user_id", id, "error", err)
		return err
	}

	slog.Debug("User ID validation passed", "user_id", id)
	return nil
}

func validateName(name string) error {
	slog.Debug("Validating name", "name", name, "length", len(name))

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
	slog.Debug("Validating email", "email", email, "length", len(email))

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
	slog.Debug("Validating age", "age", age)

	if age <= 0 || age >= 150 {
		return ErrAgeInvalid
	}
	return nil
}

func validateUserID(id string) error {
	slog.Debug("Validating user ID format", "user_id", id, "length", len(id))

	if strings.TrimSpace(id) == "" {
		return ErrUserIDRequired
	}
	return nil
}
