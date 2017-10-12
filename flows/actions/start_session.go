package actions

import (
	"encoding/json"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
)

// TypeStartSession is the type for the start session action
const TypeStartSession string = "start_session"

// StartSessionAction can be used to trigger sessions for other contacts and groups
//
// ```
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "start_session",
//     "flow": {"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d", "name": "Registration"},
//     "groups": [
//       {"uuid": "8f8e2cae-3c8d-4dce-9c4b-19514437e427", "name": "New contacts"}
//     ]
//   }
// ```
//
// @action start_session
type StartSessionAction struct {
	BaseAction
	Flow     *flows.FlowReference      `json:"flow" validate:"required"`
	Contacts []*flows.ContactReference `json:"contacts,omitempty" validate:"dive"`
	Groups   []*flows.GroupReference   `json:"groups,omitempty" validate:"dive"`
}

// Type returns the type of this action
func (a *StartSessionAction) Type() string { return TypeStartSession }

// Validate validates our action is valid
func (a *StartSessionAction) Validate(assets flows.SessionAssets) error {
	if _, err := assets.GetFlow(a.Flow.UUID); err != nil {
		return err
	}
	for _, g := range a.Groups {
		if _, err := assets.GetGroup(g.UUID); err != nil {
			return err
		}
	}
	return nil
}

// Execute runs our action
func (a *StartSessionAction) Execute(run flows.FlowRun, step flows.Step) ([]flows.Event, error) {
	runSnapshot, err := json.Marshal(run.Snapshot())
	if err != nil {
		return nil, err
	}

	return []flows.Event{events.NewSessionTriggeredEvent(a.Flow, a.Contacts, a.Groups, runSnapshot)}, nil
}