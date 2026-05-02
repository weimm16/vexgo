package api

import (
	"encoding/json"
	"strconv"
)

// JSONUint unmarshals from both JSON number and JSON string.
type JSONUint struct {
	Value uint
	Valid bool
}

func (j *JSONUint) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == `""` {
		j.Valid = false
		return nil
	}
	// Try number first
	if val, err := strconv.ParseUint(str, 10, 64); err == nil {
		j.Value = uint(val)
		j.Valid = true
		return nil
	}
	// Try quoted string (strip quotes)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	if val, err := strconv.ParseUint(str, 10, 64); err == nil {
		j.Value = uint(val)
		j.Valid = true
		return nil
	}
	return &json.UnmarshalTypeError{Value: "string/number", Type: nil}
}

func (j JSONUint) MarshalJSON() ([]byte, error) {
	if !j.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatUint(uint64(j.Value), 10)), nil
}
