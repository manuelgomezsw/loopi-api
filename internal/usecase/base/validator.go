package base

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field   string
	Rule    string
	Message string
	Params  []interface{}
}

// Validator provides validation utilities for use cases
type Validator struct {
	rules map[string][]ValidationRule
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// ValidateID validates an ID parameter
func (v *Validator) ValidateID(id int) error {
	if id <= 0 {
		return errors.New("ID must be a positive integer")
	}
	return nil
}

// ValidateEntity validates an entity using reflection and tags
func (v *Validator) ValidateEntity(entity interface{}) error {
	if entity == nil {
		return errors.New("entity cannot be nil")
	}

	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() == reflect.Ptr {
		if entityValue.IsNil() {
			return errors.New("entity cannot be nil")
		}
		entityValue = entityValue.Elem()
	}

	entityType := entityValue.Type()

	for i := 0; i < entityValue.NumField(); i++ {
		field := entityValue.Field(i)
		fieldType := entityType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get validation tags
		jsonTag := fieldType.Tag.Get("json")
		gormTag := fieldType.Tag.Get("gorm")
		validateTag := fieldType.Tag.Get("validate")

		// Skip fields with json:"-"
		if jsonTag == "-" {
			continue
		}

		fieldName := getFieldName(jsonTag, fieldType.Name)

		// Validate required fields
		if isRequired(gormTag, validateTag) {
			if err := v.validateRequired(field, fieldName); err != nil {
				return err
			}
		}

		// Validate field based on type and tags
		if err := v.validateField(field, fieldName, validateTag, gormTag); err != nil {
			return err
		}
	}

	return nil
}

// ValidateString validates a string field
func (v *Validator) ValidateString(value, fieldName string, rules ...string) error {
	for _, rule := range rules {
		switch {
		case rule == "required":
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("%s is required", fieldName)
			}
		case strings.HasPrefix(rule, "min:"):
			minLen := extractNumber(rule)
			if len(value) < minLen {
				return fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
			}
		case strings.HasPrefix(rule, "max:"):
			maxLen := extractNumber(rule)
			if len(value) > maxLen {
				return fmt.Errorf("%s must be at most %d characters", fieldName, maxLen)
			}
		case rule == "email":
			if !isValidEmail(value) {
				return fmt.Errorf("%s must be a valid email address", fieldName)
			}
		case rule == "alpha":
			if !isAlpha(value) {
				return fmt.Errorf("%s must contain only letters", fieldName)
			}
		case rule == "alphanumeric":
			if !isAlphanumeric(value) {
				return fmt.Errorf("%s must contain only letters and numbers", fieldName)
			}
		}
	}
	return nil
}

// ValidateNumber validates a numeric field
func (v *Validator) ValidateNumber(value interface{}, fieldName string, rules ...string) error {
	for _, rule := range rules {
		switch {
		case rule == "positive":
			if !isPositive(value) {
				return fmt.Errorf("%s must be positive", fieldName)
			}
		case rule == "non_negative":
			if !isNonNegative(value) {
				return fmt.Errorf("%s must be non-negative", fieldName)
			}
		case strings.HasPrefix(rule, "min:"):
			min := extractNumber(rule)
			if !isGreaterThanOrEqual(value, min) {
				return fmt.Errorf("%s must be at least %d", fieldName, min)
			}
		case strings.HasPrefix(rule, "max:"):
			max := extractNumber(rule)
			if !isLessThanOrEqual(value, max) {
				return fmt.Errorf("%s must be at most %d", fieldName, max)
			}
		}
	}
	return nil
}

// ValidateUpdateFields validates fields for update operations
func (v *Validator) ValidateUpdateFields(fields map[string]interface{}, allowedFields []string) (map[string]interface{}, error) {
	if len(fields) == 0 {
		return nil, errors.New("no fields provided for update")
	}

	// Create allowed fields map for O(1) lookup
	allowed := make(map[string]bool)
	for _, field := range allowedFields {
		allowed[field] = true
	}

	cleaned := make(map[string]interface{})
	for field, value := range fields {
		if !allowed[field] {
			continue // Skip disallowed fields
		}

		// Validate non-empty values
		if value == nil || value == "" {
			continue // Skip empty values
		}

		cleaned[field] = value
	}

	if len(cleaned) == 0 {
		return nil, errors.New("no valid fields provided for update")
	}

	return cleaned, nil
}

// Helper functions
func (v *Validator) validateRequired(field reflect.Value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		if strings.TrimSpace(field.String()) == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Bool:
		// Bool fields are always valid
	case reflect.Ptr, reflect.Interface:
		if field.IsNil() {
			return fmt.Errorf("%s is required", fieldName)
		}
	}
	return nil
}

func (v *Validator) validateField(field reflect.Value, fieldName, validateTag, gormTag string) error {
	if validateTag == "" {
		return nil
	}

	rules := strings.Split(validateTag, ",")

	switch field.Kind() {
	case reflect.String:
		return v.ValidateString(field.String(), fieldName, rules...)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.ValidateNumber(field.Int(), fieldName, rules...)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.ValidateNumber(field.Uint(), fieldName, rules...)
	case reflect.Float32, reflect.Float64:
		return v.ValidateNumber(field.Float(), fieldName, rules...)
	}

	return nil
}

func getFieldName(jsonTag, fieldName string) string {
	if jsonTag != "" && !strings.Contains(jsonTag, ",") {
		return jsonTag
	}
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	return strings.ToLower(fieldName)
}

func isRequired(gormTag, validateTag string) bool {
	return strings.Contains(gormTag, "not null") || strings.Contains(validateTag, "required")
}

func extractNumber(rule string) int {
	parts := strings.Split(rule, ":")
	if len(parts) != 2 {
		return 0
	}

	// Dynamic number extraction using strconv.Atoi
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0 // Return 0 if conversion fails
	}
	return number
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isAlpha(str string) bool {
	alphaRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)
	return alphaRegex.MatchString(str)
}

func isAlphanumeric(str string) bool {
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)
	return alphanumericRegex.MatchString(str)
}

func isPositive(value interface{}) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() > 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() > 0
	case float32, float64:
		return reflect.ValueOf(v).Float() > 0
	}
	return false
}

func isNonNegative(value interface{}) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() >= 0
	case uint, uint8, uint16, uint32, uint64:
		return true // uint is always non-negative
	case float32, float64:
		return reflect.ValueOf(v).Float() >= 0
	}
	return false
}

func isGreaterThanOrEqual(value interface{}, min int) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() >= int64(min)
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() >= uint64(min)
	case float32, float64:
		return reflect.ValueOf(v).Float() >= float64(min)
	}
	return false
}

func isLessThanOrEqual(value interface{}, max int) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() <= int64(max)
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() <= uint64(max)
	case float32, float64:
		return reflect.ValueOf(v).Float() <= float64(max)
	}
	return false
}
