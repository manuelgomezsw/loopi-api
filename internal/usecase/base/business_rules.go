package base

import (
	"fmt"
	"reflect"
)

// BusinessRule represents a business rule
type BusinessRule struct {
	Name        string
	Description string
	Validate    func(entity interface{}, context map[string]interface{}) error
}

// BusinessRuleEngine manages and executes business rules
type BusinessRuleEngine struct {
	rules  map[string][]BusinessRule
	logger *Logger
}

// NewBusinessRuleEngine creates a new business rule engine
func NewBusinessRuleEngine(logger *Logger) *BusinessRuleEngine {
	return &BusinessRuleEngine{
		rules:  make(map[string][]BusinessRule),
		logger: logger,
	}
}

// RegisterRule registers a business rule for an entity type
func (bre *BusinessRuleEngine) RegisterRule(entityType string, rule BusinessRule) {
	if bre.rules[entityType] == nil {
		bre.rules[entityType] = make([]BusinessRule, 0)
	}
	bre.rules[entityType] = append(bre.rules[entityType], rule)
}

// ValidateEntity validates an entity against all registered rules
func (bre *BusinessRuleEngine) ValidateEntity(entity interface{}, context map[string]interface{}) error {
	entityType := getEntityTypeName(entity)

	rules, exists := bre.rules[entityType]
	if !exists {
		return nil // No rules registered for this entity type
	}

	for _, rule := range rules {
		bre.logger.LogBusinessRule("ValidateEntity", rule.Name, "executing", map[string]interface{}{
			"entity_type": entityType,
			"rule_desc":   rule.Description,
		})

		if err := rule.Validate(entity, context); err != nil {
			bre.logger.LogBusinessRule("ValidateEntity", rule.Name, "violated", map[string]interface{}{
				"entity_type": entityType,
				"error":       err.Error(),
			})
			return fmt.Errorf("business rule '%s' violated: %w", rule.Name, err)
		}

		bre.logger.LogBusinessRule("ValidateEntity", rule.Name, "passed", map[string]interface{}{
			"entity_type": entityType,
		})
	}

	return nil
}

// ValidateOperation validates an operation against specific rules
func (bre *BusinessRuleEngine) ValidateOperation(operation string, entity interface{}, context map[string]interface{}) error {
	entityType := getEntityTypeName(entity)
	operationKey := fmt.Sprintf("%s.%s", entityType, operation)

	rules, exists := bre.rules[operationKey]
	if !exists {
		return nil // No operation-specific rules
	}

	for _, rule := range rules {
		if err := rule.Validate(entity, context); err != nil {
			bre.logger.LogBusinessRule("ValidateOperation", rule.Name, "violated", map[string]interface{}{
				"operation":   operation,
				"entity_type": entityType,
				"error":       err.Error(),
			})
			return fmt.Errorf("operation rule '%s' violated: %w", rule.Name, err)
		}
	}

	return nil
}

// Common business rules factory functions

// CreateRequiredFieldRule creates a rule that validates required fields
func CreateRequiredFieldRule(fieldName string) BusinessRule {
	return BusinessRule{
		Name:        fmt.Sprintf("required_%s", fieldName),
		Description: fmt.Sprintf("Field %s is required", fieldName),
		Validate: func(entity interface{}, context map[string]interface{}) error {
			entityValue := reflect.ValueOf(entity)
			if entityValue.Kind() == reflect.Ptr {
				entityValue = entityValue.Elem()
			}

			field := entityValue.FieldByName(fieldName)
			if !field.IsValid() {
				return fmt.Errorf("field %s not found", fieldName)
			}

			if isEmptyValue(field) {
				return fmt.Errorf("field %s is required", fieldName)
			}

			return nil
		},
	}
}

// CreateUniqueFieldRule creates a rule that validates field uniqueness
func CreateUniqueFieldRule(fieldName string, checkUnique func(value interface{}) (bool, error)) BusinessRule {
	return BusinessRule{
		Name:        fmt.Sprintf("unique_%s", fieldName),
		Description: fmt.Sprintf("Field %s must be unique", fieldName),
		Validate: func(entity interface{}, context map[string]interface{}) error {
			entityValue := reflect.ValueOf(entity)
			if entityValue.Kind() == reflect.Ptr {
				entityValue = entityValue.Elem()
			}

			field := entityValue.FieldByName(fieldName)
			if !field.IsValid() {
				return fmt.Errorf("field %s not found", fieldName)
			}

			isUnique, err := checkUnique(field.Interface())
			if err != nil {
				return fmt.Errorf("error checking uniqueness of %s: %w", fieldName, err)
			}

			if !isUnique {
				return fmt.Errorf("field %s must be unique", fieldName)
			}

			return nil
		},
	}
}

// CreateRangeRule creates a rule that validates numeric ranges
func CreateRangeRule(fieldName string, min, max float64) BusinessRule {
	return BusinessRule{
		Name:        fmt.Sprintf("range_%s", fieldName),
		Description: fmt.Sprintf("Field %s must be between %v and %v", fieldName, min, max),
		Validate: func(entity interface{}, context map[string]interface{}) error {
			entityValue := reflect.ValueOf(entity)
			if entityValue.Kind() == reflect.Ptr {
				entityValue = entityValue.Elem()
			}

			field := entityValue.FieldByName(fieldName)
			if !field.IsValid() {
				return fmt.Errorf("field %s not found", fieldName)
			}

			var value float64
			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value = float64(field.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value = float64(field.Uint())
			case reflect.Float32, reflect.Float64:
				value = field.Float()
			default:
				return fmt.Errorf("field %s is not numeric", fieldName)
			}

			if value < min || value > max {
				return fmt.Errorf("field %s must be between %v and %v", fieldName, min, max)
			}

			return nil
		},
	}
}

// CreateConditionalRule creates a rule that depends on other field values
func CreateConditionalRule(name string, description string, condition func(entity interface{}) bool, rule func(entity interface{}) error) BusinessRule {
	return BusinessRule{
		Name:        name,
		Description: description,
		Validate: func(entity interface{}, context map[string]interface{}) error {
			if condition(entity) {
				return rule(entity)
			}
			return nil
		},
	}
}

// CreateContextRule creates a rule that depends on context
func CreateContextRule(name string, description string, rule func(entity interface{}, context map[string]interface{}) error) BusinessRule {
	return BusinessRule{
		Name:        name,
		Description: description,
		Validate:    rule,
	}
}

// Helper functions
func getEntityTypeName(entity interface{}) string {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	return entityType.Name()
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// Common business rule patterns

// StoreBusinessRules returns common business rules for stores
func StoreBusinessRules() []BusinessRule {
	return []BusinessRule{
		CreateRequiredFieldRule("Name"),
		CreateRequiredFieldRule("Code"),
		CreateRangeRule("FranchiseID", 1, 999999),
		{
			Name:        "valid_store_code_format",
			Description: "Store code must be exactly 3 characters",
			Validate: func(entity interface{}, context map[string]interface{}) error {
				entityValue := reflect.ValueOf(entity)
				if entityValue.Kind() == reflect.Ptr {
					entityValue = entityValue.Elem()
				}

				field := entityValue.FieldByName("Code")
				if !field.IsValid() {
					return fmt.Errorf("Code field not found")
				}

				code := field.String()
				if len(code) != 3 {
					return fmt.Errorf("store code must be exactly 3 characters")
				}

				return nil
			},
		},
	}
}

// UserBusinessRules returns common business rules for users
func UserBusinessRules() []BusinessRule {
	return []BusinessRule{
		CreateRequiredFieldRule("FirstName"),
		CreateRequiredFieldRule("LastName"),
		CreateRequiredFieldRule("Email"),
		{
			Name:        "valid_email_format",
			Description: "Email must be in valid format",
			Validate: func(entity interface{}, context map[string]interface{}) error {
				entityValue := reflect.ValueOf(entity)
				if entityValue.Kind() == reflect.Ptr {
					entityValue = entityValue.Elem()
				}

				field := entityValue.FieldByName("Email")
				if !field.IsValid() {
					return fmt.Errorf("Email field not found")
				}

				email := field.String()
				// Simple email validation (could be improved)
				if len(email) < 5 || !contains(email, "@") || !contains(email, ".") {
					return fmt.Errorf("email must be in valid format")
				}

				return nil
			},
		},
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
