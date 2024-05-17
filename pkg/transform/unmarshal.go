package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"go.uber.org/multierr"
)

type FieldError struct {
	field         string
	err           error
	fieldAffected bool
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("field: %s, error: %v", e.field, e.err)
}

func (e *FieldError) Field() string {
	return e.field
}

func (e *FieldError) IsFieldAffected() bool {
	return e.fieldAffected
}

// Unmarshal converts data to a struct.
// It can replace json.Unmarshal with greater precision and fewer conversion errors of numeric and alphanumeric types.
// Data can be a 'map', 'string', or 'byte slice'. Result should be a pointer to a struct to be filled.
// This function replaces the old 'DataToStruct' function.
func Unmarshal(data interface{}, result interface{}) error {
	switch v := data.(type) {
	case []byte:
		return UnmarshalBytes(v, result)
	case string:
		return UnmarshalString(v, result)
	case map[string]interface{}:
		return UnmarshalMap(v, result)
	case []interface{}:
		return UnmarshalSlice(v, result)
	case []map[string]interface{}:
		return UnmarshalSliceOfMaps(v, result)
	default:
		return fmt.Errorf("unsupported data type: %s", reflect.TypeOf(data))
	}
}

func UnmarshalBytes(data []byte, result interface{}) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err == nil {
		return MapToStruct(raw, result)
	}
	var rawSlice []map[string]interface{}
	err = json.Unmarshal(data, &rawSlice)
	if err != nil {
		return err
	}
	rawInterfaceSlice := make([]interface{}, len(rawSlice))
	for i, item := range rawSlice {
		rawInterfaceSlice[i] = item
	}
	return Unmarshal(rawInterfaceSlice, result)
}

func UnmarshalString(data string, result interface{}) error {
	var raw map[string]interface{}
	err := json.Unmarshal([]byte(data), &raw)
	if err != nil {
		return err
	}
	return MapToStruct(raw, result)
}

func UnmarshalMap(data map[string]interface{}, result interface{}) error {
	return MapToStruct(data, result)
}

func UnmarshalSlice(data []interface{}, result interface{}) error {
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result should be a pointer to a slice")
	}
	resultValue = resultValue.Elem()
	for _, item := range data {
		newItem := reflect.New(resultValue.Type().Elem()).Interface()
		err := Unmarshal(item, newItem)
		if err != nil {
			return err
		}
		resultValue = reflect.Append(resultValue, reflect.ValueOf(newItem).Elem())
	}
	reflect.ValueOf(result).Elem().Set(resultValue)
	return nil
}

func UnmarshalSliceOfMaps(data []map[string]interface{}, result interface{}) error {
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result should be a pointer to a slice")
	}
	resultValue = resultValue.Elem()
	for _, item := range data {
		newItem := reflect.New(resultValue.Type().Elem()).Interface()
		err := Unmarshal(item, newItem)
		if err != nil {
			return err
		}
		resultValue = reflect.Append(resultValue, reflect.ValueOf(newItem).Elem())
	}
	reflect.ValueOf(result).Elem().Set(resultValue)
	return nil
}

func MapToStruct(data map[string]interface{}, result interface{}) error {
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("result should be a pointer to a struct")
	}
	resultValue = resultValue.Elem()

	var multiErr error

	for key, value := range data {
		field := resultValue.FieldByName(key)
		if !field.IsValid() {
			continue
		}
		if !field.CanSet() {
			continue
		}

		var convertedValue reflect.Value
		var err error
		switch field.Kind() {
		case reflect.String:
			var strVal string
			strVal, err = ConvertString(value)
			convertedValue = reflect.ValueOf(strVal)
		case reflect.Int:
			var intVal int
			intVal, err = ConvertInt(value)
			convertedValue = reflect.ValueOf(intVal)
		case reflect.Float64:
			var floatVal float64
			floatVal, err = ConvertFloat64(value)
			convertedValue = reflect.ValueOf(floatVal)
		case reflect.Bool:
			var boolVal bool
			boolVal, err = ConvertBool(value)
			convertedValue = reflect.ValueOf(boolVal)
		case reflect.Slice:
			var sliceVal []interface{}
			sliceVal, err = ConvertSlice(value)
			convertedValue = reflect.ValueOf(sliceVal)
		default:
			return fmt.Errorf("unsupported field type: %s", field.Type())
		}

		if err != nil {
			multiErr = multierr.Append(multiErr, &FieldError{field: key, err: err, fieldAffected: true})
		} else {
			field.Set(convertedValue)
		}
	}

	return multiErr
}

func ConvertString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(v), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

func ConvertInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case string:
		return strconv.Atoi(v)
	case float64:
		return int(v), nil
	default:
		return strconv.Atoi(fmt.Sprintf("%v", value))
	}
}

func ConvertFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseFloat(v, 64)
	case int:
		return float64(v), nil
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	}
}

func ConvertBool(value interface{}) (bool, error) {
	boolVal, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("cannot convert value: %v to bool", value)
	}
	return boolVal, nil
}

func ConvertSlice(value interface{}) ([]interface{}, error) {
	sliceVal, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert value: %v to slice", value)
	}
	return sliceVal, nil
}
