package modifiers_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/assets/static"
	"github.com/nyaruka/goflow/envs"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/actions/modifiers"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/test"
	"github.com/nyaruka/goflow/utils/dates"
	"github.com/nyaruka/goflow/utils/jsonx"
	"github.com/nyaruka/goflow/utils/uuids"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModifierTypes(t *testing.T) {
	env := envs.NewBuilder().Build()
	assets, err := test.LoadSessionAssets(env, "testdata/_assets.json")
	require.NoError(t, err)

	for typeName := range modifiers.RegisteredTypes {
		testModifierType(t, assets, typeName)
	}
}

func testModifierType(t *testing.T, sessionAssets flows.SessionAssets, typeName string) {
	testPath := fmt.Sprintf("testdata/%s.json", typeName)
	testFile, err := ioutil.ReadFile(testPath)
	require.NoError(t, err)

	tests := []struct {
		Description   string          `json:"description"`
		ContactBefore json.RawMessage `json:"contact_before"`
		Modifier      json.RawMessage `json:"modifier"`

		ContactAfter json.RawMessage `json:"contact_after"`
		Events       json.RawMessage `json:"events"`
	}{}

	err = jsonx.Unmarshal(testFile, &tests)
	require.NoError(t, err)

	defer dates.SetNowSource(dates.DefaultNowSource)
	defer uuids.SetGenerator(uuids.DefaultGenerator)

	for i, tc := range tests {
		dates.SetNowSource(dates.NewFixedNowSource(time.Date(2018, 10, 18, 14, 20, 30, 123456, time.UTC)))
		uuids.SetGenerator(uuids.NewSeededGenerator(12345))

		testName := fmt.Sprintf("test '%s' for modifier type '%s'", tc.Description, typeName)

		// read the modifier to be tested
		modifier, err := modifiers.ReadModifier(sessionAssets, tc.Modifier, assets.PanicOnMissing)
		require.NoError(t, err, "error loading modifier in %s", testName)
		assert.Equal(t, typeName, modifier.Type())

		// read the initial contact state
		contact, err := flows.ReadContact(sessionAssets, tc.ContactBefore, assets.PanicOnMissing)
		require.NoError(t, err, "error loading contact_before in %s", testName)

		// apply the modifier
		eventLog := test.NewEventLog()
		modifier.Apply(envs.NewBuilder().Build(), sessionAssets, contact, eventLog.Log)

		// clone test case and populate with actual values
		actual := tc

		// re-marshal the modifier
		actual.Modifier, err = jsonx.Marshal(modifier)
		require.NoError(t, err)

		// and the contact
		actual.ContactAfter, _ = jsonx.Marshal(contact)

		// and the events
		actual.Events, _ = jsonx.Marshal(eventLog.Events)

		if !test.UpdateSnapshots {
			// check the modifier marshaled correctly
			test.AssertEqualJSON(t, tc.Modifier, actual.Modifier, "marshal mismatch in %s", testName)

			// check contact is in the expected state
			test.AssertEqualJSON(t, tc.ContactAfter, actual.ContactAfter, "contact mismatch in %s", testName)

			// check events are what we expected
			test.AssertEqualJSON(t, tc.Events, actual.Events, "events mismatch in %s", testName)
		} else {
			tests[i] = actual
		}
	}

	if test.UpdateSnapshots {
		actualJSON, err := jsonx.MarshalPretty(tests)
		require.NoError(t, err)

		err = ioutil.WriteFile(testPath, actualJSON, 0666)
		require.NoError(t, err)
	}
}

func TestConstructors(t *testing.T) {
	env := envs.NewBuilder().Build()
	assets, err := test.LoadSessionAssets(env, "testdata/_assets.json")
	require.NoError(t, err)

	nexmo := assets.Channels().Get("3a05eaf5-cb1b-4246-bef1-f277419c83a7")
	age := assets.Fields().Get("age")
	ageValue := types.NewXNumberFromInt(37)
	testers := assets.Groups().Get("b7cf0d83-f1c9-411c-96fd-c511a4cfa86d")
	la, _ := time.LoadLocation("America/Los_Angeles")

	tests := []struct {
		modifier flows.Modifier
		json     string
	}{
		{
			modifiers.NewChannel(nexmo),
			`{
				"type": "channel",
				"channel": {
					"uuid": "3a05eaf5-cb1b-4246-bef1-f277419c83a7",
					"name": "Nexmo"
				}
			}`,
		},
		{
			modifiers.NewField(age, flows.NewValue(types.NewXText("37 years"), nil, &ageValue, "", "", "")),
			`{
				"type": "field",
				"field": {
					"key": "age",
					"name": "Age"
				},
				"value": {
					"text": "37 years",
					"number": 37
				}
			}`,
		},
		{
			modifiers.NewGroups([]*flows.Group{testers}, modifiers.GroupsAdd),
			`{
				"type": "groups",
				"groups": [
					{
						"uuid": "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d",
						"name": "Testers"
					}
				],
				"modification": "add"
			}`,
		},
		{
			modifiers.NewLanguage(envs.Language("fra")),
			`{
				"type": "language",
				"language": "fra"
			}`,
		},
		{
			modifiers.NewName("Bob"),
			`{
				"type": "name",
				"name": "Bob"
			}`,
		},
		{
			modifiers.NewIsBlocked(modifiers.ShouldBlock),
			`{
				"type": "is_blocked",
				"modification": "block"
			}`,
		},
		{
			modifiers.NewIsStopped(modifiers.ShouldStop),
			`{
				"type": "is_stopped",
				"modification": "stop"
			}`,
		},
		{
			modifiers.NewIsStopped(modifiers.ShouldUnstop),
			`{
				"type": "is_stopped",
				"modification": "unstop"
			}`,
		},
		{
			modifiers.NewIsBlocked(modifiers.ShouldUnblock),
			`{
				"type": "is_blocked",
				"modification": "unblock"
			}`,
		},
		{
			modifiers.NewTimezone(la),
			`{
				"type": "timezone",
				"timezone": "America/Los_Angeles"
			}`,
		},
		{
			modifiers.NewURN(urns.URN("tel:+1234567890"), modifiers.URNAppend),
			`{
				"type": "urn",
				"urn": "tel:+1234567890",
				"modification": "append"
			}`,
		},
	}

	for _, tc := range tests {
		modifierJSON, err := jsonx.Marshal(tc.modifier)
		require.NoError(t, err)
		test.AssertEqualJSON(t, []byte(tc.json), modifierJSON, "marshal mismatch for modifier %s", string(modifierJSON))
	}
}

func TestReadModifier(t *testing.T) {
	env := envs.NewBuilder().Build()
	missingAssets := make([]assets.Reference, 0)
	missing := func(a assets.Reference, err error) { missingAssets = append(missingAssets, a) }

	sessionAssets, err := engine.NewSessionAssets(env, static.NewEmptySource(), nil)
	require.NoError(t, err)

	// error if no type field
	_, err = modifiers.ReadModifier(sessionAssets, []byte(`{"foo": "bar"}`), missing)
	assert.EqualError(t, err, "field 'type' is required")

	// error if we don't recognize the type
	_, err = modifiers.ReadModifier(sessionAssets, []byte(`{"type": "do_the_foo", "foo": "bar"}`), missing)
	assert.EqualError(t, err, "unknown type: 'do_the_foo'")

	// no-modifier error and a missing asset record if we load a channel modifier for a channel that no longer exists
	mod, err := modifiers.ReadModifier(sessionAssets, []byte(`{"type": "channel", "channel": {"uuid": "8632b9f0-ac2f-40ad-808f-77781a444dc9", "name": "Nexmo"}}`), missing)
	assert.Equal(t, modifiers.ErrNoModifier, err)
	assert.Nil(t, mod)
	assert.Equal(t, assets.NewChannelReference(assets.ChannelUUID("8632b9f0-ac2f-40ad-808f-77781a444dc9"), "Nexmo"), missingAssets[len(missingAssets)-1])

	// no-modifier error and a missing asset record if we load a field modifier for a field that no longer exists
	mod, err = modifiers.ReadModifier(sessionAssets, []byte(`{"type": "field", "field": {"key": "gender", "name": "Gender"}, "value": {"text": "M"}}`), missing)
	assert.Equal(t, modifiers.ErrNoModifier, err)
	assert.Nil(t, mod)
	assert.Equal(t, assets.NewFieldReference("gender", "Gender"), missingAssets[len(missingAssets)-1])

	// no-modifier error if we load a groups modifier and none of its groups exist
	mod, err = modifiers.ReadModifier(sessionAssets, []byte(`{"type": "groups", "modification": "add", "groups": [{"uuid": "8632b9f0-ac2f-40ad-808f-77781a444dc9", "name": "Testers"}]}`), missing)
	assert.Equal(t, modifiers.ErrNoModifier, err)
	assert.Nil(t, mod)
	assert.Equal(t, assets.NewGroupReference(assets.GroupUUID("8632b9f0-ac2f-40ad-808f-77781a444dc9"), "Testers"), missingAssets[len(missingAssets)-1])

	// but if at least one of its groups exists, we still get a modifier
	source, _ := static.NewSource([]byte(`{
		"groups": [
			{"uuid": "4349cdd6-5385-46f3-8e55-5750dd4f35fb", "name": "Winners"}
		]
	}`))
	sessionAssets, err = engine.NewSessionAssets(env, source, nil)
	require.NoError(t, err)

	mod, err = modifiers.ReadModifier(sessionAssets, []byte(`{"type": "groups", "modification": "add", "groups": [{"uuid": "cd1a2aa6-0d9d-4a8c-b32d-ca5de9c43bdb", "name": "Losers"}, {"uuid": "4349cdd6-5385-46f3-8e55-5750dd4f35fb", "name": "Winners"}]}`), missing)
	assert.NoError(t, err)
	assert.NotNil(t, mod)
	assert.Equal(t, "groups", mod.Type())
	assert.Equal(t, assets.NewGroupReference(assets.GroupUUID("cd1a2aa6-0d9d-4a8c-b32d-ca5de9c43bdb"), "Losers"), missingAssets[len(missingAssets)-1])
}
