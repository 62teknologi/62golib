package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	"github.com/iancoleman/strcase"
)

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func ResponseData(status string, message string, data any) *response {
	return &response{
		Status:  status,
		Message: message,
		Data:    data,
	}
}

func ResponseDataPaginate(status string, message string, data any, pagination, filter map[string]any, search map[string]any, summary map[string]any) map[string]any {
	responses := map[string]any{
		"status":  status,
		"message": message,
		"data":    data,
	}

	if len(pagination) != 0 {
		responses["pagination"] = pagination
	}

	if len(filter) != 0 {
		responses["filter"] = filter
	}

	if len(summary) != 0 {
		responses["summary"] = summary
	}

	if len(search) != 0 {
		responses["search"] = search
	}

	return responses
}

func JsonFileParser(fileDir string) (map[string]any, error) {
	jsonFile, err := os.Open(fileDir)

	if err != nil {
		return nil, fmt.Errorf("error while parsing input: %w", err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var input map[string]any
	json.Unmarshal(byteValue, &input)

	return input, nil
}

func MultiMapValuesShifter(transformer map[string]any, values []map[string]any) []map[string]any {
	var customResponses []map[string]any

	for _, value := range values {
		transformerC := map[string]any{}

		for k, v := range transformer {
			transformerC[k] = v
		}

		MapValuesShifter(transformerC, value)
		AttachBelongsTo(transformerC, value)
		AttachOperation(transformerC, value)

		customResponses = append(customResponses, transformerC)
	}

	return customResponses
}

func MapValuesShifter(dest map[string]any, origin map[string]any) map[string]any {
	val := reflect.ValueOf(dest)

	for _, inp := range val.MapKeys() {
		key := inp.String()
		if origin[key] == nil {
			if origin[strcase.ToCamel(key)] != nil {
				dest[key] = origin[strcase.ToCamel(key)]
				continue
			}
			continue
		}
		dest[key] = origin[key]
	}

	return dest
}

func MapNullValuesRemover(m map[string]any) {
	val := reflect.ValueOf(m)
	for _, e := range val.MapKeys() {
		v := val.MapIndex(e)
		if v.IsNil() || len(fmt.Sprintf("%v", v)) == 0 {
			delete(m, e.String())
			continue
		}
		switch t := v.Interface().(type) {
		// If key is a JSON object (Go Map), use recursion to go deeper
		case map[string]any:
			MapNullValuesRemover(t)
		}
	}
}

func LogJson(data any) {
	bytes, err := json.Marshal(data)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(bytes))
}

func Prepare1toM(key string, id any, value any) (valuesC []map[string]any) {
	values := value.([]any)
	valuesC = make([]map[string]any, len(values))

	for i, v := range values {
		valuesC[i] = v.(map[string]any)
		valuesC[i][key] = id
	}

	return valuesC
}

func PrepareMtoM(key1 string, id any, key2 any, value any) (valuesC []map[string]any) {
	values := value.([]any)
	valuesC = make([]map[string]any, len(values))

	for i, v := range values {
		valuesC[i] = map[string]any{
			key1:          id,
			key2.(string): int(v.(float64)),
		}
	}

	return valuesC
}

func convertAnyToString(anySlice []any) []string {
	strSlice := make([]string, len(anySlice))

	for i, v := range anySlice {
		strSlice[i] = fmt.Sprintf("%v", v)
	}

	return strSlice
}

func filterSliceByMapIndex(data []map[string]any, index string, productID any) []any {
	var values []any

	for _, item := range data {
		if item[index] == productID {
			values = append(values, item)
		}
	}

	return values
}

func ConvertToInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		num, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return num
	default:
		return 0
	}
}
