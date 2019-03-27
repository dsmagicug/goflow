package flows

import (
	"fmt"
	"strings"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/utils"
)

// Template represents messaging templates used by channels types such as WhatsApp
type Template struct {
	assets.Template
}

// NewTemplate returns a new template objects based on the passed in asset
func NewTemplate(t assets.Template) *Template {
	return &Template{Template: t}
}

// FindTranslation finds the matching translation for the passed in channel and languages (in priority order)
func (t *Template) FindTranslation(channel assets.ChannelUUID, langs []utils.Language) TemplateContent {
	// first iterate through and find all translations that are for this channel
	matches := make(map[utils.Language]assets.TemplateTranslation)
	for _, tr := range t.Template.Translations() {
		if tr.Channel().UUID == channel {
			matches[tr.Language()] = tr
		}
	}

	// now find the first that matches our language
	for _, lang := range langs {
		tr := matches[lang]
		if tr != nil {
			return TemplateContent(tr.Content())
		}
	}

	return NilTemplateContent
}

// Asset returns the underlying asset
func (t *Template) Asset() assets.Template { return t.Template }

// NilTemplateContent is our constant for nil content
const NilTemplateContent = TemplateContent("")

// TemplateContent represents the translated content for a template
type TemplateContent string

// Substitute substitutes the passed in variables in our template
func (c TemplateContent) Substitute(vars []string) string {
	s := string(c)
	for i, v := range vars {
		s = strings.ReplaceAll(s, fmt.Sprintf("{{%d}}", i), v)
	}

	return s
}

// TemplateAssets is our type for all the templates in an environment
type TemplateAssets struct {
	templates []*Template
	byName    map[string]*Template
}

// NewTemplateAssets creates a new template list
func NewTemplateAssets(ts []assets.Template) *TemplateAssets {
	templates := make([]*Template, len(ts))
	byName := make(map[string]*Template)
	for i, t := range ts {
		template := NewTemplate(t)
		templates[i] = template
		byName[t.Name()] = template
	}

	return &TemplateAssets{
		templates: templates,
		byName:    byName,
	}
}

// FindTranslation looks through our list of templates to find the template matching the passed in name
// If no template or translation is found then empty string is returned
func (l *TemplateAssets) FindTranslation(name string, channel *assets.ChannelReference, langs []utils.Language) TemplateContent {
	// no channel, can't match to a template
	if channel == nil {
		return ""
	}

	template := l.byName[name]

	// not found, no template
	if template == nil {
		return ""
	}

	// look through our translations looking for a match by both channel and translation
	return template.FindTranslation(channel.UUID, langs)
}
