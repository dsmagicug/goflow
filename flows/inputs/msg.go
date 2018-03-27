package inputs

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/utils"
)

// TypeMsg is a constant for incoming messages
const TypeMsg string = "msg"

// MsgInput is a message which can be used as input
type MsgInput struct {
	baseInput
	urn         urns.URN
	text        string
	attachments []flows.Attachment
}

// NewMsgInput creates a new user input based on a message
func NewMsgInput(uuid flows.InputUUID, channel flows.Channel, createdOn time.Time, urn urns.URN, text string, attachments []flows.Attachment) *MsgInput {
	return &MsgInput{
		baseInput:   baseInput{uuid: uuid, channel: channel, createdOn: createdOn},
		urn:         urn,
		text:        text,
		attachments: attachments,
	}
}

// Type returns the type of this event
func (i *MsgInput) Type() string { return TypeMsg }

// Resolve resolves the given key when this input is referenced in an expression
func (i *MsgInput) Resolve(key string) interface{} {
	switch key {
	case "urn":
		return i.urn
	case "text":
		return i.text
	case "attachments":
		return i.attachments
	}
	return i.baseInput.Resolve(key)
}

// String returns our default value if evaluated in a context, our text in our case
func (i *MsgInput) String() string {
	var parts []string
	if i.text != "" {
		parts = append(parts, i.text)
	}
	for _, attachment := range i.attachments {
		parts = append(parts, attachment.URL())
	}
	return strings.Join(parts, "\n")
}

var _ flows.Input = (*MsgInput)(nil)

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

type msgInputEnvelope struct {
	baseInputEnvelope
	URN         urns.URN           `json:"urn" validate:"urn"`
	Text        string             `json:"text" validate:"required"`
	Attachments []flows.Attachment `json:"attachments,omitempty"`
}

func ReadMsgInput(session flows.Session, data json.RawMessage) (*MsgInput, error) {
	input := MsgInput{}
	i := msgInputEnvelope{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	err = utils.Validate(i)
	if err != nil {
		return nil, err
	}

	// lookup the channel
	var channel flows.Channel
	if i.Channel != nil {
		channel, err = session.Assets().GetChannel(i.Channel.UUID)
		if err != nil {
			return nil, err
		}
	}

	input.baseInput.uuid = i.UUID
	input.baseInput.channel = channel
	input.baseInput.createdOn = i.CreatedOn
	input.urn = i.URN
	input.text = i.Text
	input.attachments = i.Attachments
	return &input, nil
}

// MarshalJSON marshals this msg input into JSON
func (i *MsgInput) MarshalJSON() ([]byte, error) {
	var envelope msgInputEnvelope

	if i.Channel() != nil {
		envelope.baseInputEnvelope.Channel = i.Channel().Reference()
	}
	envelope.baseInputEnvelope.UUID = i.UUID()
	envelope.baseInputEnvelope.CreatedOn = i.CreatedOn()
	envelope.URN = i.urn
	envelope.Text = i.text
	envelope.Attachments = i.attachments

	return json.Marshal(envelope)
}
