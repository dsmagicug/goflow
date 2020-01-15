package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/nyaruka/goflow/excellent/tools"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/test"
)

func TestExplorableContext(t *testing.T) {
	context := types.NewXObject(map[string]types.XValue{
		"__default__": types.NewXText("Bob"),
		"foo": types.NewXArray(
			types.NewXObject(map[string]types.XValue{
				"__default__": types.NewXText("Bob"),
				"deprecated":  types.NewXText("hide me"),
				"bar":         types.NewXNumberFromInt(123),
			}),
			types.NewXObject(map[string]types.XValue{
				"bar": types.NewXNumberFromInt(345),
			}),
		),
		"deprecated": types.NewXText("hide me"),
		"bar":        types.NewXNumberFromInt(256),
	})

	explorable := tools.ExplorableContext(context, [][]string{{"deprecated"}, {"foo", "0", "deprecated"}})
	explorableJSON, _ := json.Marshal(explorable)

	// check that defaults are included but excluded paths aren't
	test.AssertEqualJSON(t, []byte(`{
		"__default__": "Bob",
		"foo": [
			{
				"__default__": "Bob",
				"bar": 123
			},
			{
				"bar": 345
			}
		],
		"bar": 256
	}`), explorableJSON, "explorable context mismatch")
}
