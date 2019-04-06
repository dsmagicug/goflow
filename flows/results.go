package flows

import (
	"encoding/json"
	"time"

	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/utils"
)

// Result describes a value captured during a run's execution. It might have been implicitly created by a router, or explicitly
// created by a [set_run_result](#action:set_run_result) action.It renders as its value in a template, and has the following
// properties which can be accessed:
//
//  * `value` the value of the result
//  * `category` the category of the result
//  * `category_localized` the localized category of the result
//  * `input` the input associated with the result
//  * `node_uuid` the UUID of the node where the result was created
//  * `created_on` the time when the result was created
//
// Examples:
//
//   @run.results.favorite_color -> {category: Red, category_localized: Red, created_on: 2018-04-11T18:24:30.123456Z, extra: , input: , name: Favorite Color, node_uuid: f5bb9b7a-7b5e-45c3-8f0e-61b4e95edf03, value: red}
//   @run.results.favorite_color.value -> red
//   @run.results.favorite_color.category -> Red
//
// @context result
type Result struct {
	Name              string          `json:"name"`
	Value             string          `json:"value"`
	Category          string          `json:"category,omitempty"`
	CategoryLocalized string          `json:"category_localized,omitempty"`
	NodeUUID          NodeUUID        `json:"node_uuid"`
	Input             string          `json:"input,omitempty"`
	Extra             json.RawMessage `json:"extra,omitempty"`
	CreatedOn         time.Time       `json:"created_on"`
}

// NewResult creates a new result
func NewResult(name string, value string, category string, categoryLocalized string, nodeUUID NodeUUID, input string, extra json.RawMessage, createdOn time.Time) *Result {
	return &Result{
		Name:              name,
		Value:             value,
		Category:          category,
		CategoryLocalized: categoryLocalized,
		NodeUUID:          nodeUUID,
		Input:             input,
		Extra:             extra,
		CreatedOn:         createdOn,
	}
}

// ToXValue returns a representation of this object for use in expressions
func (r *Result) ToXValue(env utils.Environment) types.XValue {
	categoryLocalized := r.CategoryLocalized
	if categoryLocalized == "" {
		categoryLocalized = r.Category
	}

	return types.NewXDict(map[string]types.XValue{
		"name":               types.NewXText(r.Name),
		"value":              types.NewXText(r.Value),
		"category":           types.NewXText(r.Category),
		"category_localized": types.NewXText(categoryLocalized),
		"input":              types.NewXText(r.Input),
		"extra":              types.JSONToXValue(r.Extra),
		"node_uuid":          types.NewXText(string(r.NodeUUID)),
		"created_on":         types.NewXDateTime(r.CreatedOn),
	})
}

// Results is our wrapper around a map of snakified result names to result objects
type Results map[string]*Result

// NewResults creates a new empty set of results
func NewResults() Results {
	return make(Results, 0)
}

// Clone returns a clone of this results set
func (r Results) Clone() Results {
	clone := make(Results, len(r))
	for k, v := range r {
		clone[k] = v
	}
	return clone
}

// Save saves a new result in our map. The key is saved in a snakified format
func (r Results) Save(result *Result) {
	r[utils.Snakify(result.Name)] = result
}

// Get returns the result with the given key
func (r Results) Get(key string) *Result {
	return r[key]
}

// ToXValue returns a representation of this object for use in expressions
func (r Results) ToXValue(env utils.Environment) types.XDict {
	results := types.NewEmptyXDict()
	for k, v := range r {
		results.Put(k, v.ToXValue(env))
	}
	return results
}

// ToSimpleXDict returns a simplifed representation of this object for use in expressions
func (r Results) ToSimpleXDict(env utils.Environment) types.XDict {
	results := types.NewEmptyXDict()
	for k, v := range r {
		results.Put(k, types.NewXDict(map[string]types.XValue{
			"value":    types.NewXText(v.Value),
			"category": types.NewXText(v.Category),
		}))
	}
	return results
}
