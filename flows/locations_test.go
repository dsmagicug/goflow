package flows_test

import (
	"encoding/json"
	"testing"

	"github.com/nyaruka/goflow/flows"

	"github.com/stretchr/testify/assert"
)

var locationHierarchyJSON = `
{
	"id": "2342",
	"name": "Rwanda",
	"aliases": ["Ruanda"],		
	"children": [
		{
			"id": "234521",
			"name": "Kigali City",
			"aliases": ["Kigali", "Kigari"],
			"children": [
				{
					"id": "57735322",
					"name": "Gasabo",
					"children": [
						{
							"id": "575743222",
							"name": "Gisozi"
						},
						{
							"id": "457378732",
							"name": "Ndera"
						}
					]
				},
				{
					"id": "46547322",
					"name": "Nyarugenge",
					"children": []
				}
			]
		},
		{
			"id": "347535",
			"name": "Eastern Province"
		}
	]
}`

func TestLocationHierarchy(t *testing.T) {
	hierarchy, err := flows.ReadLocationHierarchy(json.RawMessage(locationHierarchyJSON))
	assert.NoError(t, err)

	rwanda := hierarchy.Root()
	assert.Equal(t, flows.LocationID("2342"), rwanda.ID())
	assert.Equal(t, flows.LocationLevel(0), rwanda.Level())
	assert.Equal(t, "Rwanda", rwanda.Name())
	assert.Equal(t, []string{"Ruanda"}, rwanda.Aliases())
	assert.Nil(t, rwanda.Parent())
	assert.Equal(t, 2, len(rwanda.Children()))

	kigali := rwanda.Children()[0]
	assert.Equal(t, flows.LocationID("234521"), kigali.ID())
	assert.Equal(t, flows.LocationLevel(1), kigali.Level())
	assert.Equal(t, "Kigali City", kigali.Name())
	assert.Equal(t, []string{"Kigali", "Kigari"}, kigali.Aliases())
	assert.Equal(t, rwanda, kigali.Parent())
	assert.Equal(t, 2, len(kigali.Children()))

	gasabo := kigali.Children()[0]
	assert.Equal(t, flows.LocationID("57735322"), gasabo.ID())
	assert.Equal(t, flows.LocationLevel(2), gasabo.Level())
	assert.Equal(t, "Gasabo", gasabo.Name())
	assert.Equal(t, kigali, gasabo.Parent())
	assert.Equal(t, 2, len(gasabo.Children()))

	ndera := gasabo.Children()[1]
	assert.Equal(t, flows.LocationID("457378732"), ndera.ID())
	assert.Equal(t, flows.LocationLevel(3), ndera.Level())
	assert.Equal(t, "Ndera", ndera.Name())
	assert.Equal(t, gasabo, ndera.Parent())
	assert.Equal(t, 0, len(ndera.Children()))

	assert.Equal(t, kigali, hierarchy.FindByID(flows.LocationID("234521"), flows.LocationLevel(1)))
	assert.Equal(t, gasabo, hierarchy.FindByID(flows.LocationID("57735322"), flows.LocationLevel(2)))

	assert.Nil(t, hierarchy.FindByID(flows.LocationID("xxxxx"), flows.LocationLevel(1)))  // no such ID
	assert.Nil(t, hierarchy.FindByID(flows.LocationID("234521"), flows.LocationLevel(8))) // no such level
	assert.Nil(t, hierarchy.FindByID(flows.LocationID("234521"), flows.LocationLevel(2))) // wrong level

	assert.Equal(t, []*flows.Location{kigali}, hierarchy.FindByName("kigari", flows.LocationLevel(1), nil))
	assert.Equal(t, []*flows.Location{kigali}, hierarchy.FindByName("kigari", flows.LocationLevel(1), rwanda))
	assert.Equal(t, []*flows.Location{gasabo}, hierarchy.FindByName("GASABO", flows.LocationLevel(2), nil))
	assert.Equal(t, []*flows.Location{gasabo}, hierarchy.FindByName("GASABO", flows.LocationLevel(2), kigali))

	assert.Equal(t, []*flows.Location{}, hierarchy.FindByName("boston", flows.LocationLevel(1), nil))    // no such name
	assert.Equal(t, []*flows.Location{}, hierarchy.FindByName("kigari", flows.LocationLevel(8), nil))    // no such level
	assert.Equal(t, []*flows.Location{}, hierarchy.FindByName("kigari", flows.LocationLevel(2), nil))    // wrong level
	assert.Equal(t, []*flows.Location{}, hierarchy.FindByName("kigari", flows.LocationLevel(2), gasabo)) // wrong parent
}