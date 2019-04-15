package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nyaruka/goflow/utils"
)

// XDict is a dictionary of keys and values.
//
//   @(dict("foo", 1, "bar", "x")) -> {bar: x, foo: 1}
//   @(dict("foo", 1, "bar", "x").bar) -> x
//   @(dict("foo", 1, "bar", "x")["bar"]) -> x
//   @(count(dict("foo", 1, "bar", "x"))) -> 2
//   @(json(dict("foo", 1, "bar", "x"))) -> {"bar":"x","foo":1}
//
// @type dict
type XDict struct {
	XValue
	XCountable

	data   map[string]XValue
	source func() map[string]XValue
}

// NewXDict returns a new dict with the given items
func NewXDict(data map[string]XValue) *XDict {
	return &XDict{
		data: data,
	}
}

// NewXLazyDict returns a new lazy dict with the source function
func NewXLazyDict(source func() map[string]XValue) *XDict {
	return &XDict{
		source: source,
	}
}

// Describe returns a representation of this type for error messages
func (x *XDict) Describe() string { return "dict" }

// ToXText converts this type to text
func (x *XDict) ToXText(env utils.Environment) XText {
	pairs := make([]string, 0, x.Count())
	for _, k := range x.keys(true) {
		vAsText, xerr := ToXText(env, x.values()[k])
		if xerr != nil {
			vAsText = xerr.ToXText(env)
		}

		pairs = append(pairs, fmt.Sprintf("%s: %s", k, vAsText.Native()))
	}
	return NewXText("{" + strings.Join(pairs, ", ") + "}")
}

// ToXBoolean converts this type to a bool
func (x *XDict) ToXBoolean() XBoolean {
	return NewXBoolean(x.Count() > 0)
}

// MarshalJSON converts this type to internal JSON
func (x *XDict) MarshalJSON() ([]byte, error) {
	marshaled := make(map[string]json.RawMessage, x.Count())
	for k, v := range x.values() {
		asJSON, err := ToXJSON(v)
		if err == nil {
			marshaled[k] = json.RawMessage(asJSON.Native())
		}
	}
	return json.Marshal(marshaled)
}

// Count is called when the length of this object is requested in an expression
func (x *XDict) Count() int {
	return len(x.values())
}

// Get retrieves the named item from this dict
func (x *XDict) Get(key string) (XValue, bool) {
	key = strings.ToLower(key)
	for k, v := range x.values() {
		if strings.ToLower(k) == key {
			return v, true
		}
	}

	return nil, false
}

// Keys returns the keys of this dict
func (x *XDict) Keys() []string {
	return x.keys(false)
}

// String returns the native string representation of this type for debugging
func (x *XDict) String() string {
	pairs := make([]string, 0, x.Count())
	for _, k := range x.keys(true) {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, String(x.values()[k])))
	}
	return "XDict{" + strings.Join(pairs, ", ") + "}"
}

// Equals determines equality for this type
func (x *XDict) Equals(other *XDict) bool {
	keys1 := x.keys(true)
	keys2 := other.keys(true)

	if len(keys1) != len(keys2) {
		return false
	}

	for k, key := range keys1 {
		if key != keys2[k] {
			return false
		}

		if !Equals(x.values()[key], other.values()[key]) {
			return false
		}
	}

	return true
}

func (x *XDict) keys(sorted bool) []string {
	keys := make([]string, 0, x.Count())
	for key := range x.values() {
		keys = append(keys, key)
	}
	if sorted {
		sort.Strings(keys)
	}
	return keys
}

func (x *XDict) values() map[string]XValue {
	if x.data == nil {
		x.data = x.source()
	}
	return x.data
}

// XDictEmpty is the empty dict
var XDictEmpty = NewXDict(map[string]XValue{})

var _ json.Marshaler = (*XDict)(nil)

// ToXDict converts the given value to a dict
func ToXDict(env utils.Environment, x XValue) (*XDict, XError) {
	if utils.IsNil(x) {
		return XDictEmpty, nil
	}
	if IsXError(x) {
		return XDictEmpty, x.(XError)
	}

	asDict, isDict := x.(*XDict)
	if isDict {
		return asDict, nil
	}

	return XDictEmpty, NewXErrorf("unable to convert %s to a dict", Describe(x))
}
