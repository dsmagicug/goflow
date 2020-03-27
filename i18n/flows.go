package i18n

import (
	"errors"
	"fmt"
	"net/url"
	"sort"

	"github.com/nyaruka/goflow/envs"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/utils"
	"github.com/nyaruka/goflow/utils/dates"
	"github.com/nyaruka/goflow/utils/uuids"
)

// describes the location of a piece of extracted text
type textLocation struct {
	Flow     flows.Flow
	UUID     uuids.UUID
	Property string
	Index    int
}

type extractedText struct {
	Locations   []textLocation
	Base        string
	Translation string
	Unique      bool
}

func getBaseLanguage(set []flows.Flow) envs.Language {
	if len(set) == 0 {
		return envs.NilLanguage
	}
	baseLanguage := set[0].Language()
	for _, flow := range set[1:] {
		if baseLanguage != flow.Language() {
			return envs.NilLanguage
		}
	}
	return baseLanguage
}

// ExtractFromFlows extracts a PO file from a set of flows
func ExtractFromFlows(initialComment string, translationsLanguage envs.Language, excludeProperties []string, sources ...flows.Flow) (*PO, error) {
	// check all flows have same base language
	baseLanguage := getBaseLanguage(sources)
	if baseLanguage == envs.NilLanguage {
		return nil, errors.New("can't extract from flows with differing base languages")
	} else if translationsLanguage == baseLanguage {
		translationsLanguage = envs.NilLanguage // we'll create a POT in the base language (i.e. no translations)
	}

	extracted := extractFromFlows(translationsLanguage, excludeProperties, sources)

	merged := mergeExtracted(extracted)

	return poFromExtracted(initialComment, translationsLanguage, merged), nil
}

func extractFromFlows(lang envs.Language, excludeProperties []string, sources []flows.Flow) []*extractedText {
	exclude := utils.StringSet(excludeProperties)
	extracted := make([]*extractedText, 0)

	for _, flow := range sources {
		var targetTranslation flows.Translation
		if lang != envs.NilLanguage {
			targetTranslation = flow.Localization().GetTranslation(lang)
		}

		for _, node := range flow.Nodes() {
			node.EnumerateLocalizables(func(uuid uuids.UUID, property string, texts []string) {
				if !exclude[property] {
					exts := extractFromProperty(flow, uuid, property, texts, targetTranslation)
					extracted = append(extracted, exts...)
				}
			})
		}
	}

	return extracted
}

func extractFromProperty(flow flows.Flow, uuid uuids.UUID, property string, texts []string, targetTranslation flows.Translation) []*extractedText {
	extracted := make([]*extractedText, 0)

	// look up target translation if we have one
	targets := make([]string, len(texts))
	if targetTranslation != nil {
		translation := targetTranslation.GetTextArray(uuid, property)
		if translation != nil {
			for t := range targets {
				if t < len(translation) {
					targets[t] = translation[t]
				}
			}
		}
	}

	for t, text := range texts {
		if text != "" {
			extracted = append(extracted, &extractedText{
				Locations: []textLocation{
					textLocation{
						Flow:     flow,
						UUID:     uuid,
						Property: property,
						Index:    t,
					},
				},
				Base:        text,
				Translation: targets[t],
				Unique:      false,
			})
		}
	}

	return extracted
}

func mergeExtracted(extracted []*extractedText) []*extractedText {
	// organize extracted texts by their base text
	byBase := make(map[string][]*extractedText)
	for _, e := range extracted {
		byBase[e.Base] = append(byBase[e.Base], e)
	}

	// get the list of unique base text values and sort A-Z
	bases := make([]string, 0, len(byBase))
	for b := range byBase {
		bases = append(bases, b)
	}
	sort.Strings(bases)

	merged := make([]*extractedText, 0)

	for _, base := range bases {
		extractionsForBase := byBase[base]

		majorityTranslation := majorityTranslation(extractionsForBase)

		// all extractions with majority translation or no translation get merged into a new context-less extraction
		mergedLocations := make([]textLocation, 0)

		for _, ext := range extractionsForBase {
			if ext.Translation == majorityTranslation || ext.Translation == "" {
				mergedLocations = append(mergedLocations, ext.Locations[0])
			} else {
				merged = append(merged, ext)
			}
		}

		merged = append(merged, &extractedText{
			Locations:   mergedLocations,
			Base:        base,
			Translation: majorityTranslation,
			Unique:      true,
		})
	}

	return merged
}

// finds the majority non-empty translation
func majorityTranslation(extracted []*extractedText) string {
	counts := make(map[string]int)
	for _, e := range extracted {
		if e.Translation != "" {
			counts[e.Translation]++
		}
	}
	max := 0
	majority := ""
	for _, e := range extracted {
		if counts[e.Translation] > max {
			majority = e.Translation
			max = counts[e.Translation]
		}
	}
	return majority
}

func poFromExtracted(initialComment string, lang envs.Language, extracted []*extractedText) *PO {
	header := NewPOHeader(initialComment, dates.Now(), lang.ToISO639_2(envs.NilCountry))
	po := NewPO(header)

	for _, ext := range extracted {
		references := make([]string, len(ext.Locations))
		for i, loc := range ext.Locations {
			flowName := url.QueryEscape(loc.Flow.Name())
			references[i] = fmt.Sprintf("%s/%s/%s:%d", flowName, string(loc.UUID), loc.Property, loc.Index)
		}
		sort.Strings(references)

		context := ""
		if !ext.Unique {
			context = fmt.Sprintf("%s/%s:%d", string(ext.Locations[0].UUID), ext.Locations[0].Property, ext.Locations[0].Index)
		}

		entry := &POEntry{
			Comment: POComment{
				References: references,
			},
			MsgContext: context,
			MsgID:      ext.Base,
			MsgStr:     ext.Translation,
		}

		po.AddEntry(entry)
	}

	return po
}