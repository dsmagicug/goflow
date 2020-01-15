package tools

import (
	"strconv"

	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/utils"
)

// ExplorableContext transforms the given context for use as documentation. It can exclude certain paths
// and make explicit defaults render like regular properties.
func ExplorableContext(context *types.XObject, excludePaths [][]string) map[string]interface{} {
	include := func(path []string) bool {
		for _, p := range excludePaths {
			if utils.StringSliceEquals(path, p) {
				return false
			}
		}
		return true
	}

	cloned := cloneContext(context, nil, include, true)

	return cloned.(map[string]interface{})
}

func cloneContext(v types.XValue, path []string, include func([]string) bool, defaults bool) interface{} {
	switch typed := v.(type) {
	case *types.XObject:
		vals := make(map[string]interface{}, len(typed.Properties())+1)
		if defaults && !utils.IsNil(typed.Default()) && typed.Default() != typed {
			vals["__default__"] = typed.Default()
		}
		for _, p := range typed.Properties() {
			c, _ := typed.Get(p)
			newPath := append(path, p)
			if include(newPath) {
				vals[p] = cloneContext(c, newPath, include, defaults)
			}
		}
		return vals
	case *types.XArray:
		vals := make([]interface{}, 0, typed.Count())
		for i := 0; i < typed.Count(); i++ {
			c := typed.Get(i)
			newPath := append(path, strconv.Itoa(i))
			if include(newPath) {
				vals = append(vals, cloneContext(c, newPath, include, defaults))
			}
		}
		return vals
	default:
		return typed
	}
}
