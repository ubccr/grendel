// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
)

func elems(names ...string) *gpb.Path {
	path := &gpb.Path{}
	for _, name := range names {
		path.Elem = append(path.Elem, &gpb.PathElem{Name: name})
	}
	return path
}

type leaf struct {
	elems []*gpb.PathElem
	value any
}

func (l leaf) str() string {
	switch v := l.value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

// stripModule drops the YANG module prefix JSON_IETF puts on names and identityref values, e.g. openconfig-lldp-types:MAC_ADDRESS. Neither can hold a colon of its own, but free text such as a description can, so it is never passed here.
func stripModule(s string) string {
	if _, name, ok := strings.Cut(s, ":"); ok {
		return name
	}
	return s
}

func leavesFromNotifications(notifs []*gpb.Notification) ([]leaf, error) {
	var out []leaf
	for _, n := range notifs {
		var prefix []*gpb.PathElem
		if n.GetPrefix() != nil {
			prefix = n.GetPrefix().GetElem()
		}
		for _, u := range n.GetUpdate() {
			elems := append(append([]*gpb.PathElem{}, prefix...), u.GetPath().GetElem()...)
			for _, e := range elems {
				e.Name = stripModule(e.Name)
			}
			val, isJSON, err := decodeValue(u.GetVal())
			if err != nil {
				return nil, fmt.Errorf("decoding %s: %w", pathString(elems), err)
			}
			if isJSON {
				out = flatten(out, elems, val)
			} else {
				out = append(out, leaf{elems: elems, value: val})
			}
		}
	}
	return out, nil
}

func decodeValue(v *gpb.TypedValue) (any, bool, error) {
	var raw []byte
	switch t := v.GetValue().(type) {
	case *gpb.TypedValue_JsonIetfVal:
		raw = t.JsonIetfVal
	case *gpb.TypedValue_JsonVal:
		raw = t.JsonVal
	case *gpb.TypedValue_AsciiVal:
		return t.AsciiVal, false, nil
	default:
		// a type ToScalar cannot read, such as any_val, is left empty so one odd leaf doesn't fail the whole response
		scalar, _ := value.ToScalar(v)
		return scalar, false, nil
	}
	// JSON is decoded here rather than by ToScalar, which reads numbers as float64 and so is inexact past 2^53
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// flatten walks a JSON subtree into leaves. A list item's scalar fields become the keys of its path element, since JSON carries list keys as ordinary fields of each item.
func flatten(out []leaf, elems []*gpb.PathElem, v any) []leaf {
	switch t := v.(type) {
	case map[string]any:
		names := make([]string, 0, len(t))
		for k := range t {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			child := t[k]
			name := stripModule(k)
			if items, ok := child.([]any); ok && len(items) > 0 {
				if _, isObj := items[0].(map[string]any); isObj {
					for _, item := range items {
						obj, _ := item.(map[string]any)
						elem := &gpb.PathElem{Name: name, Key: map[string]string{}}
						for ik, iv := range obj {
							switch iv.(type) {
							case map[string]any, []any:
							default:
								elem.Key[stripModule(ik)] = leaf{value: iv}.str()
							}
						}
						out = flatten(out, appendElem(elems, elem), obj)
					}
					continue
				}
			}
			out = flatten(out, appendElem(elems, &gpb.PathElem{Name: name}), child)
		}
		return out
	case []any:
		// A list returned whole at its own path, e.g. a Get on /interfaces/interface.
		if len(elems) > 0 && len(t) > 0 {
			if _, isObj := t[0].(map[string]any); isObj {
				parent, last := elems[:len(elems)-1], elems[len(elems)-1]
				return flatten(out, parent, map[string]any{last.Name: t})
			}
		}
		return append(out, leaf{elems: elems, value: v})
	default:
		return append(out, leaf{elems: elems, value: v})
	}
}

// appendElem copies so sibling branches never share a backing array.
func appendElem(elems []*gpb.PathElem, e *gpb.PathElem) []*gpb.PathElem {
	out := make([]*gpb.PathElem, len(elems)+1)
	copy(out, elems)
	out[len(elems)] = e
	return out
}

func pathString(elems []*gpb.PathElem) string {
	var b strings.Builder
	for _, e := range elems {
		b.WriteString("/")
		b.WriteString(e.Name)
		keys := make([]string, 0, len(e.Key))
		for k := range e.Key {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "[%s=%s]", k, e.Key[k])
		}
	}
	return b.String()
}
