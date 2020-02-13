package flows_test

import (
	"encoding/json"
	"testing"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/assets/static"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/actions"
	"github.com/nyaruka/goflow/flows/definition"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/flows/routers"
	"github.com/nyaruka/goflow/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unknownAssetType struct{}

func (t *unknownAssetType) String() string   { return "x" }
func (t *unknownAssetType) Type() string     { return "unknown" }
func (t *unknownAssetType) Identity() string { return "unknown[]" }
func (t *unknownAssetType) Variable() bool   { return false }

func TestDependencies(t *testing.T) {
	action1 := actions.NewSendMsg("ed08e6b9-ed22-4294-9871-c7ac7d82cbd5", "Hi there", nil, nil, false)
	node1 := definition.NewNode("91b20e13-d6e2-42a9-b74f-bce85c9da8c8", []flows.Action{action1}, nil, nil)
	router2 := routers.NewRandom(nil, "", nil)
	node2 := definition.NewNode("7c959933-4c30-4277-9810-adc95a459bd0", nil, router2, nil)

	refs := []flows.ExtractedReference{
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewChannelReference("8286545d-d1a1-4eff-a3ad-a11ddf4bb20a", "Android")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewClassifierReference("2138cddc-118a-49ae-b290-98e03ad0573b", "Booking")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: flows.NewContactReference("0b099519-0889-4c74-b744-9122272f346a", "Bob")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewFieldReference("gender", "Gender")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewFlowReference("4f932672-7995-47f0-96e6-faf5abd2d81d", "Registration")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewGlobalReference("org_name", "Org Name")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewGroupReference("46057a92-6580-4e93-af36-2bb9c9d61e51", "Testers")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewGroupReference("377c3101-a7fc-47b1-9136-980348e362c0", "Customers")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewLabelReference("31c06b7c-010d-4f91-9590-d3fbdc2fb7ac", "Spam")},
		flows.ExtractedReference{Node: node1, Action: action1, Reference: assets.NewTemplateReference("ff958d30-f50e-48ab-a524-37ed1e9620d9", "Welcome")},
		flows.ExtractedReference{Node: node2, Router: router2, Reference: assets.NewGlobalReference("org_name", "Org Name")},
	}

	// if our assets only includes a single group, the other assets should be reported as missing
	source, err := static.NewSource([]byte(`{
			"groups": [
				{
					"uuid": "377c3101-a7fc-47b1-9136-980348e362c0",
					"name": "Customers"
				}
			]
		}`))
	require.NoError(t, err)

	sa, err := engine.NewSessionAssets(source, nil)
	require.NoError(t, err)

	deps := flows.NewDependencies(refs, sa)
	depsJSON, _ := json.Marshal(deps)
	test.AssertEqualJSON(t, []byte(`[
		{
			"missing": true,
			"name": "Android",
			"type": "channel",
			"uuid": "8286545d-d1a1-4eff-a3ad-a11ddf4bb20a"
		},
		{
			"missing": true,
			"name": "Booking",
			"type": "classifier",
			"uuid": "2138cddc-118a-49ae-b290-98e03ad0573b"
		},
		{
			"name": "Bob",
			"type": "contact",
			"uuid": "0b099519-0889-4c74-b744-9122272f346a"
		},
		{
			"key": "gender",
			"missing": true,
			"name": "Gender",
			"type": "field"
		},
		{
			"missing": true,
			"name": "Registration",
			"type": "flow",
			"uuid": "4f932672-7995-47f0-96e6-faf5abd2d81d"
		},
		{
			"key": "org_name",
			"missing": true,
			"name": "Org Name",
			"type": "global"
		},
		{
			"missing": true,
			"name": "Testers",
			"type": "group",
			"uuid": "46057a92-6580-4e93-af36-2bb9c9d61e51"
		},
		{
			"name": "Customers",
			"type": "group",
			"uuid": "377c3101-a7fc-47b1-9136-980348e362c0"
		},
		{
			"missing": true,
			"name": "Spam",
			"type": "label",
			"uuid": "31c06b7c-010d-4f91-9590-d3fbdc2fb7ac"
		},
		{
			"missing": true,
			"name": "Welcome",
			"type": "template",
			"uuid": "ff958d30-f50e-48ab-a524-37ed1e9620d9"
		}
	]`), depsJSON, "deps JSON mismatch")

	// panic if we get a dependency type we don't recognize
	assert.Panics(t, func() {
		flows.NewDependencies([]flows.ExtractedReference{
			flows.ExtractedReference{Node: node1, Action: action1, Reference: &unknownAssetType{}},
		}, sa)
	})
}

func TestNewResultSpecs(t *testing.T) {
	assert.Equal(t, []*flows.ResultSpec{}, flows.NewResultSpecs(nil))

	node1 := definition.NewNode(
		flows.NodeUUID("1fb823c3-599a-41e9-b59b-658266af3466"),
		nil,
		nil,
		[]flows.Exit{definition.NewExit(flows.ExitUUID("3c158842-24f3-4a40-bea4-7522952c0131"), "")},
	)
	node2 := definition.NewNode(
		flows.NodeUUID("0ba673a3-63b3-46f9-9246-9c727cf2917f"),
		nil,
		nil,
		[]flows.Exit{definition.NewExit(flows.ExitUUID("434ac29c-abe6-4bd7-b29b-740d517b6bb5"), "")},
	)

	extracted := []flows.ExtractedResult{
		flows.ExtractedResult{Node: node1, Info: flows.NewResultInfo("Response 1", []string{"Red", "Green"})},
		flows.ExtractedResult{Node: node1, Info: flows.NewResultInfo("Response-1", nil)},
		flows.ExtractedResult{Node: node2, Info: flows.NewResultInfo("Response-1", []string{"Green", "Blue"})},
		flows.ExtractedResult{Node: node2, Info: flows.NewResultInfo("Favorite Beer", []string{})},
	}

	specs := flows.NewResultSpecs(extracted)
	specsJSON, _ := json.Marshal(specs)

	test.AssertEqualJSON(t, []byte(`[
		{
			"key": "response_1",
			"name": "Response 1",
			"categories": [
				"Red",
				"Green",
				"Blue"
			],
			"node_uuids": [
				"1fb823c3-599a-41e9-b59b-658266af3466",
				"0ba673a3-63b3-46f9-9246-9c727cf2917f"
			]
		},
		{
			"key": "favorite_beer",
			"name": "Favorite Beer",
			"categories": [],
			"node_uuids": [
				"0ba673a3-63b3-46f9-9246-9c727cf2917f"
			]
		}
	]`), specsJSON, "result specs JSON mismatch")

	assert.Equal(t, `key=response_1|name=Response 1|categories=Red,Green`, flows.NewResultInfo("Response 1", []string{"Red", "Green"}).String())
}
