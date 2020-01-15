package tools

import (
	"github.com/nyaruka/goflow/excellent/types"
)

// ContextWalk traverses the given context invoking the callback for each value
func ContextWalk(context *types.XObject, callback func(types.XValue)) {
	contextWalk(context, callback)
}

// ContextFindObjects traverses the given context invoking the callback for each object found
func ContextFindObjects(context *types.XObject, callback func(*types.XObject)) {
	contextWalk(context, func(v types.XValue) {
		switch typed := v.(type) {
		case *types.XObject:
			callback(typed)
		}
	})
}

func contextWalk(v types.XValue, callback func(types.XValue)) {
	callback(v)

	switch typed := v.(type) {
	case *types.XObject:
		for _, p := range typed.Properties() {
			c, _ := typed.Get(p)
			contextWalk(c, callback)
		}
	case *types.XArray:
		for i := 0; i < typed.Count(); i++ {
			c := typed.Get(i)
			contextWalk(c, callback)
		}
	}
}
