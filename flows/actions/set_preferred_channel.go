package actions

import (
	"fmt"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
)

const TypeSetPreferredChannel string = "set_preferred_channel"

type PreferredChannelAction struct {
	BaseAction
	Channel *flows.ChannelReference `json:"channel"`
}

// Type returns the type of this action
func (a *PreferredChannelAction) Type() string { return TypeSetPreferredChannel }

// Validate validates our action is valid and has all the assets it needs
func (a *PreferredChannelAction) Validate(assets flows.SessionAssets) error {
	_, err := assets.GetChannel(a.Channel.UUID)
	return err
}

func (a *PreferredChannelAction) Execute(run flows.FlowRun, step flows.Step, log flows.EventLog) error {
	if run.Contact() == nil {
		log.Add(events.NewErrorEvent(fmt.Errorf("can't execute action in session without a contact")))
		return nil
	}

	log.Add(events.NewContactChannelChangedEvent(a.Channel))
	return nil
}
