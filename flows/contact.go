package flows

import (
	"encoding/json"
	"fmt"

	"time"

	"github.com/nyaruka/goflow/contactql"
	"github.com/nyaruka/goflow/utils"
)

// Contact represents a single contact
type Contact struct {
	uuid     ContactUUID
	name     string
	language utils.Language
	timezone *time.Location
	urns     URNList
	groups   *GroupList
	fields   FieldValues
	channel  Channel
}

// SetLanguage sets the language for this contact
func (c *Contact) SetLanguage(lang utils.Language) { c.language = lang }

// Language gets the language for this contact
func (c *Contact) Language() utils.Language { return c.language }

func (c *Contact) SetTimezone(tz *time.Location) {
	c.timezone = tz
}
func (c *Contact) Timezone() *time.Location { return c.timezone }

func (c *Contact) SetName(name string) { c.name = name }
func (c *Contact) Name() string        { return c.name }

func (c *Contact) URNs() URNList     { return c.urns }
func (c *Contact) UUID() ContactUUID { return c.uuid }

func (c *Contact) Groups() *GroupList  { return c.groups }
func (c *Contact) Fields() FieldValues { return c.fields }

func (c *Contact) Channel() Channel           { return c.channel }
func (c *Contact) SetChannel(channel Channel) { c.channel = channel }

func (c *Contact) Resolve(key string) interface{} {
	switch key {

	case "name":
		return c.name

	case "uuid":
		return c.uuid

	case "urns":
		return c.urns

	case "language":
		return string(c.language)

	case "groups":
		return c.groups

	case "fields":
		return c.fields

	case "timezone":
		return c.timezone

	case "urn":
		return c.urns

	case "channel":
		return c.channel
	}

	return fmt.Errorf("No field '%s' on contact", key)
}

// Default returns our default value in the context
func (c *Contact) Default() interface{} {
	return c
}

// String returns our string value in the context
func (c *Contact) String() string {
	return c.name
}

var _ utils.VariableResolver = (*Contact)(nil)

func (c *Contact) UpdateDynamicGroups(session Session) error {
	groups, err := session.Assets().GetGroupSet()
	if err != nil {
		return err
	}

	for _, group := range groups.All() {
		if group.IsDynamic() {
			qualifies, err := group.CheckDynamicMembership(session, c)
			if err != nil {
				return err
			}
			if qualifies {
				c.groups.Add(group)
			} else {
				c.groups.Remove(group)
			}
		}
	}

	return nil
}

func (c *Contact) ResolveQueryKey(key string) interface{} {
	if key == contactql.ImplicitKey {
		// TODO add urns
		return []string{c.name}
	} else if key == "name" {
		return c.name
	}

	// try as a URN scheme
	for _, urn := range c.urns {
		if key == string(urn.Scheme()) {
			return urn.Path()
		}
	}

	// try as a contact field
	for k, value := range c.fields {
		if key == k {
			return value.value
		}
	}

	return nil
}

var _ contactql.Queryable = (*Contact)(nil)

type ContactReference struct {
	UUID ContactUUID `json:"uuid"    validate:"required,uuid4"`
	Name string      `json:"name"`
}

func NewContactReference(uuid ContactUUID, name string) *ContactReference {
	return &ContactReference{UUID: uuid, Name: name}
}

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

type fieldValueEnvelope struct {
	Value     string    `json:"value"`
	CreatedOn time.Time `json:"created_on"`
}

type contactEnvelope struct {
	UUID        ContactUUID                     `json:"uuid" validate:"required,uuid4"`
	Name        string                          `json:"name"`
	Language    utils.Language                  `json:"language"`
	Timezone    string                          `json:"timezone"`
	URNs        URNList                         `json:"urns"`
	GroupUUIDs  []GroupUUID                     `json:"group_uuids,omitempty" validate:"dive,uuid4"`
	Fields      map[FieldKey]fieldValueEnvelope `json:"fields,omitempty"`
	ChannelUUID ChannelUUID                     `json:"channel_uuid,omitempty" validate:"omitempty,uuid4"`
}

// ReadContact decodes a contact from the passed in JSON
func ReadContact(session Session, data json.RawMessage) (*Contact, error) {
	var envelope contactEnvelope
	var err error

	err = json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, err
	}
	err = utils.Validate(envelope)
	if err != nil {
		return nil, err
	}

	c := &Contact{}
	c.uuid = envelope.UUID
	c.name = envelope.Name
	c.language = envelope.Language

	tz, err := time.LoadLocation(envelope.Timezone)
	if err != nil {
		return nil, err
	}
	c.timezone = tz

	if envelope.URNs == nil {
		c.urns = make(URNList, 0)
	} else {
		c.urns = envelope.URNs
	}

	if envelope.GroupUUIDs == nil {
		c.groups = NewGroupList([]*Group{})
	} else {
		groups := make([]*Group, len(envelope.GroupUUIDs))
		for g := range envelope.GroupUUIDs {
			if groups[g], err = session.Assets().GetGroup(envelope.GroupUUIDs[g]); err != nil {
				return nil, err
			}
		}
		c.groups = NewGroupList(groups)
	}

	if envelope.Fields == nil {
		c.fields = make(FieldValues)
	} else {
		c.fields = make(FieldValues, len(envelope.Fields))
		for fieldKey, valueEnvelope := range envelope.Fields {
			field, err := session.Assets().GetField(fieldKey)
			if err != nil {
				return nil, err
			}

			value, err := field.ParseValue(session.Environment(), valueEnvelope.Value)
			if err != nil {
				return nil, err
			}

			c.fields[field.key] = NewFieldValue(field, value, valueEnvelope.CreatedOn)
		}
	}

	if envelope.ChannelUUID != "" {
		c.channel, err = session.Assets().GetChannel(envelope.ChannelUUID)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Contact) MarshalJSON() ([]byte, error) {
	var ce contactEnvelope

	ce.Name = c.name
	ce.UUID = c.uuid
	ce.Language = c.language
	ce.URNs = c.urns
	if c.timezone != nil {
		ce.Timezone = c.timezone.String()
	}

	ce.GroupUUIDs = make([]GroupUUID, c.groups.Count())
	for g, group := range c.groups.All() {
		ce.GroupUUIDs[g] = group.UUID()
	}

	ce.Fields = make(map[FieldKey]fieldValueEnvelope, len(c.fields))
	for _, v := range c.fields {
		ce.Fields[v.field.Key()] = fieldValueEnvelope{Value: v.SerializeValue(), CreatedOn: v.createdOn}
	}

	return json.Marshal(ce)
}
