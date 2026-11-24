// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Coerce turns a value from a flag, the environment, the config file, the database or the API into the canonical Go value for the option's kind
func (o *Option) Coerce(raw any) (any, error) {
	v, err := coerce(o.Kind, raw, o)
	if err != nil {
		return nil, err
	}
	if o.Validate != nil {
		if err := o.Validate(v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func coerce(kind Kind, raw any, o *Option) (any, error) {
	switch kind {
	case KindBool:
		return toBool(raw)
	case KindInt:
		n, err := toInt(raw)
		if err != nil {
			return nil, err
		}
		if o.Min != nil && n < *o.Min {
			return nil, fmt.Errorf("must be at least %d", *o.Min)
		}
		if o.Max != nil && n > *o.Max {
			return nil, fmt.Errorf("must be at most %d", *o.Max)
		}
		return n, nil
	case KindFloat:
		return toFloat(raw)
	case KindDuration:
		return toDuration(raw)
	case KindStringList:
		return toStringList(raw)
	case KindIntList:
		return toIntList(raw)
	case KindMap:
		return toMap(raw)
	case KindGroup:
		return toGroup(raw, o)
	case KindTable:
		return toTable(raw, o)
	}

	s, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("must be text, not %T", raw)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		if o.Required {
			return nil, fmt.Errorf("must not be empty")
		}
		return s, nil
	}
	switch kind {
	case KindEnum:
		i := slices.IndexFunc(o.Enum, func(e string) bool { return strings.EqualFold(e, s) })
		if i < 0 {
			return nil, fmt.Errorf("must be one of %s", strings.Join(o.Enum, ", "))
		}
		return o.Enum[i], nil
	}
	return s, nil
}

func toBool(raw any) (bool, error) {
	switch v := raw.(type) {
	case bool:
		return v, nil
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return false, fmt.Errorf("must be true or false")
		}
		return b, nil
	}
	return false, fmt.Errorf("must be true or false, not %T", raw)
}

func toInt(raw any) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if v != math.Trunc(v) {
			return 0, fmt.Errorf("must be a whole number")
		}
		return int(v), nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("must be a whole number")
		}
		return n, nil
	}
	return 0, fmt.Errorf("must be a whole number, not %T", raw)
}

func toFloat(raw any) (float64, error) {
	switch v := raw.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, fmt.Errorf("must be a number")
		}
		return f, nil
	}
	return 0, fmt.Errorf("must be a number, not %T", raw)
}

func toDuration(raw any) (time.Duration, error) {
	switch v := raw.(type) {
	case time.Duration:
		return v, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "0" {
			return 0, nil
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("must be a duration with a unit, such as 30s or 24h")
		}
		if d < 0 {
			return 0, fmt.Errorf("must not be negative")
		}
		return d, nil
	}
	// a bare number would be read as nanoseconds, which is never what was meant
	return 0, fmt.Errorf("must be a duration with a unit, such as 30s or 24h")
}

var listSeparator = regexp.MustCompile(`[\s,]+`)

// items splits a list written as text. A flag prints a list as [a,b], the environment gives a b or a,b, and the API sends JSON.
func items(raw any) ([]any, error) {
	switch v := raw.(type) {
	case []any:
		return v, nil
	case []string:
		out := make([]any, len(v))
		for i, s := range v {
			out[i] = s
		}
		return out, nil
	case []int:
		out := make([]any, len(v))
		for i, n := range v {
			out[i] = n
		}
		return out, nil
	case string:
		s := strings.TrimSpace(v)
		var parsed []any
		if strings.HasPrefix(s, "[") && json.Unmarshal([]byte(s), &parsed) == nil {
			return parsed, nil
		}
		s = strings.Trim(s, "[]")
		out := []any{}
		for _, part := range listSeparator.Split(s, -1) {
			if part != "" {
				out = append(out, part)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("must be a list, not %T", raw)
}

func toStringList(raw any) ([]string, error) {
	list, err := items(raw)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("must be a list of text, not of %T", item)
		}
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

func toIntList(raw any) ([]int, error) {
	list, err := items(raw)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(list))
	for _, item := range list {
		n, err := toInt(item)
		if err != nil {
			return nil, fmt.Errorf("must be a list of whole numbers")
		}
		out = append(out, n)
	}
	return out, nil
}

func toMap(raw any) (map[string]any, error) {
	var m map[string]any
	switch v := raw.(type) {
	case map[string]any:
		m = v
	case string:
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, fmt.Errorf("must be a JSON object")
		}
	default:
		return nil, fmt.Errorf("must be a table of entries, not %T", raw)
	}

	// viper lowercases the keys of a config file, so the database's are too or the two would not line up
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[strings.ToLower(k)] = v
	}
	return out, nil
}

func toTable(raw any, o *Option) ([]map[string]any, error) {
	var rows []any
	switch v := raw.(type) {
	case []any:
		rows = v
	case []map[string]any:
		for _, row := range v {
			rows = append(rows, row)
		}
	case string:
		if err := json.Unmarshal([]byte(v), &rows); err != nil {
			return nil, fmt.Errorf("must be a JSON list of objects")
		}
	default:
		return nil, fmt.Errorf("must be a list of entries, not %T", raw)
	}

	out := make([]map[string]any, 0, len(rows))
	for i, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("entry %d must be a table", i+1)
		}
		clean, err := toRecord(row, o.Fields)
		if err != nil {
			return nil, fmt.Errorf("entry %d %s", i+1, err)
		}
		out = append(out, clean)
	}
	return out, nil
}

func toRecord(row map[string]any, fields []Field) (map[string]any, error) {
	clean := map[string]any{}
	for name, value := range row {
		idx := slices.IndexFunc(fields, func(f Field) bool { return strings.EqualFold(f.Name, name) })
		if idx < 0 {
			return nil, fmt.Errorf("has an unknown field %s", name)
		}
		v, err := fields[idx].coerce(value)
		if err != nil {
			return nil, fmt.Errorf("%s %s", fields[idx].Name, err)
		}
		if !empty(v) {
			clean[fields[idx].Name] = v
		}
	}
	return clean, nil
}

// empty reports whether a field's value says nothing, which is the same as leaving the field out
func empty(v any) bool {
	switch t := v.(type) {
	case string:
		return t == ""
	case []string:
		return len(t) == 0
	case []int:
		return len(t) == 0
	}
	return false
}

func missing(o *Option, record map[string]any) (string, bool) {
	for _, f := range o.Fields {
		if _, ok := record[f.Name]; f.Required && !ok {
			return f.Name, true
		}
	}
	return "", false
}

var groupName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// toGroup reads named records. A plain list, which is how they were written before they had names, is named 10, 20, 30 and so on, which keeps its order and leaves room to put a named one in between.
func toGroup(raw any, o *Option) (map[string]any, error) {
	named := map[string]any{}
	switch v := raw.(type) {
	case map[string]any:
		named = v
	case []any:
		for i, row := range v {
			named[strconv.Itoa((i+1)*10)] = row
		}
	case []map[string]any:
		for i, row := range v {
			named[strconv.Itoa((i+1)*10)] = row
		}
	default:
		return nil, fmt.Errorf("must be a table of named entries, not %T", raw)
	}

	out := make(map[string]any, len(named))
	for name, item := range named {
		name = strings.ToLower(name)
		if !groupName.MatchString(name) {
			return nil, fmt.Errorf("%q is not a name: use lowercase letters, digits, - and _", name)
		}
		row, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s must be a table", name)
		}
		clean, err := toRecord(row, o.Fields)
		if err != nil {
			return nil, fmt.Errorf("%s %s", name, err)
		}
		if field, ok := missing(o, clean); ok {
			return nil, fmt.Errorf("%s %s must be set", name, field)
		}
		out[name] = clean
	}
	return out, nil
}

// Text renders a canonical value the way it is typed in the UI or on the command line. Lists, maps and tables are written as JSON.
func Text(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case int:
		return strconv.Itoa(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case time.Duration:
		return durationText(t)
	}
	b, err := json.Marshal(storedForm(v))
	if err != nil {
		return ""
	}
	return string(b)
}

func durationText(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	return s
}

// storedForm is the JSON shape a canonical value has in the database and in an export
func storedForm(v any) any {
	switch t := v.(type) {
	case time.Duration:
		return durationText(t)
	case []string:
		if t == nil {
			return []string{}
		}
	case []int:
		if t == nil {
			return []int{}
		}
	}
	return v
}
