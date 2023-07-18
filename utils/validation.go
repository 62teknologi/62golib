// orriginally generated by chatgpt 3.5

package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Validation struct {
	Rules  map[string]any
	Data   map[string]any
	Errors map[string]map[string]string
}

func NewValidation(data map[string]any, rules map[string]any) *Validation {
	return &Validation{
		Rules:  rules,
		Data:   data,
		Errors: make(map[string]map[string]string),
	}
}

func Validate(data map[string]any, rules map[string]any) (*Validation, bool) {
	for key := range rules {
		_, ok := data[key]
		if rules[key] != nil {
			if reflect.TypeOf(rules[key]).Kind() != reflect.Map && reflect.TypeOf(rules[key]).Kind() != reflect.Slice {
				if !ok && !strings.Contains(rules[key].(string), "required") {
					delete(rules, key)
				}
			}
		}
	}

	// validate sub fields e.g. { items: [{ name: "required|max:255"}] }
	filteredSubData := make([]string, 0)
	for key, value := range data {
		if value != nil {
			if reflect.TypeOf(value).Kind() == reflect.Slice {
				filteredSubData = append(filteredSubData, key)
			}
		}
	}
	for _, key := range filteredSubData {
		subData := data[key].([]any)
		if rules[key] != nil {
			subRules := rules[key].([]any)
			if subData[0] != nil {
				if reflect.TypeOf(subData[0]).Kind() == reflect.Map {
					validation := NewValidation(subData[0].(map[string]any), subRules[0].(map[string]any))
					if validation.validate() {
						return validation, true
					} else {
						return validation, false
					}
				}
			}
		}
	}

	// Validate parent fields
	validation := NewValidation(data, rules)
	if validation.validate() {
		return validation, true
	} else {
		return validation, false
	}
}

func (v *Validation) validate() bool {
	for field, rule := range v.Rules {
		val, ok := v.Data[field]
		value := fmt.Sprintf("%v", val)

		if _, is_string := rule.(string); !is_string {
			continue
		}

		rules := strings.Split(fmt.Sprintf("%v", (rule)), "|")

		v.Errors[field] = map[string]string{}

		for _, r := range rules {
			if r == "" {
				continue
			} else if r == "required" {
				if len(value) == 0 || !ok {
					v.Errors[field]["required"] = field + " field is required"
					continue
				}
			} else if value == "" || value == "<nil>" {
				break
			} else if r == "email" {
				if !isEmailValid(value) {
					v.Errors[field]["email"] = field + " field must be a valid email"
					continue
				}
			} else if strings.HasPrefix(r, "min:") {
				length, err := parseRule(r)
				if err != nil || len(value) < length {
					v.Errors[field]["min"] = fmt.Sprintf("%s field must be at least %d characters long", field, length)
					continue
				}
			} else if strings.HasPrefix(r, "max:") {
				length, err := parseRule(r)
				if err != nil || len(value) > length {
					v.Errors[field]["max"] = fmt.Sprintf("%s field must be at most %d characters long", field, length)
					continue
				}
			} else if r == "number" {
				numberRegex := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
				if !numberRegex.MatchString(value) {
					v.Errors[field]["number"] = field + " field must be a number"
					continue
				}
			} else {
				v.Errors[field][r] = field + " field has an invalid rule"
				continue
			}
		}

		if len(v.Errors[field]) == 0 {
			delete(v.Errors, field)
		}
	}

	return len(v.Errors) != 0
}

func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	return emailRegex.MatchString(email)
}

func parseRule(rule string) (int, error) {
	lengthStr := strings.TrimPrefix(rule, "min:")
	lengthStr = strings.TrimPrefix(lengthStr, "max:")
	length, err := strconv.Atoi(lengthStr)

	if err != nil {
		return 0, err
	}

	return length, nil
}
