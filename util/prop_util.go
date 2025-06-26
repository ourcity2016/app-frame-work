package util

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func LoadProperties(filename string) (map[string]string, error) {
	config := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if shouldSkipLine(line) {
			continue
		}

		key, value, ok := parseKeyValue(line)
		if ok {
			config[key] = value
		}
	}

	return config, scanner.Err()
}

func shouldSkipLine(line string) bool {
	return len(line) == 0 || strings.HasPrefix(line, "#")
}

func parseKeyValue(line string) (string, string, bool) {
	sepIndex := strings.IndexAny(line, "=:")
	if sepIndex < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:sepIndex]), strings.TrimSpace(line[sepIndex+1:]), true
}

func MapToStruct(m map[string]string, s interface{}) error {
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if field.Kind() == reflect.Struct {
			if err := mapNestedStruct(m, field, fieldType); err != nil {
				return err
			}
		} else {
			setSimpleField(field, fieldType, m)
		}
	}
	return nil
}

func mapNestedStruct(m map[string]string, field reflect.Value, fieldType reflect.StructField) error {
	nestedType := field.Type()

	for j := 0; j < field.NumField(); j++ {
		nestedField := field.Field(j)
		nestedFieldType := nestedType.Field(j)

		if nestedField.Kind() == reflect.Struct {
			if err := mapDeepNestedStruct(m, fieldType, nestedField, nestedFieldType); err != nil {
				return err
			}
		} else {
			key := buildKey(fieldType.Name, nestedFieldType.Name)
			setFieldValue(nestedField, nestedFieldType, m, key)
		}
	}
	return nil
}

func mapDeepNestedStruct(m map[string]string, parentFieldType reflect.StructField,
	structField reflect.Value, structFieldType reflect.StructField) error {
	deepType := structField.Type()

	for k := 0; k < structField.NumField(); k++ {
		deepField := structField.Field(k)
		deepFieldType := deepType.Field(k)

		key := buildKey(parentFieldType.Name, structFieldType.Name, deepFieldType.Name)
		setFieldValue(deepField, deepFieldType, m, key)
	}
	return nil
}

func buildKey(parts ...string) string {
	return strings.Join(parts, ".")
}

func setSimpleField(field reflect.Value, fieldType reflect.StructField, m map[string]string) {
	setFieldValue(field, fieldType, m, fieldType.Name)
}

func setFieldValue(field reflect.Value, fieldType reflect.StructField, m map[string]string, key string) {
	value, exists := m[key]
	if !exists {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			field.SetBool(b)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			items := splitAndTrim(value, ",")
			field.Set(reflect.ValueOf(items))
		}
	}
}

func splitAndTrim(s, sep string) []string {
	items := strings.Split(s, sep)
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}
	return items
}
