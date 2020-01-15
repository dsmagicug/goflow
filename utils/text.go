package utils

import (
	"regexp"
	"sort"
	"strings"
)

var snakedChars = regexp.MustCompile(`[^\p{L}\d_]+`)

// treats sequences of letters/numbers/_/' as tokens, and symbols as individual tokens
var wordTokenRegex = regexp.MustCompile(`[\pM\pL\pN_']+|\pS`)

// Snakify turns the passed in string into a context reference. We replace all whitespace
// characters with _ and replace any duplicate underscores
func Snakify(text string) string {
	return strings.ToLower(snakedChars.ReplaceAllString(strings.TrimSpace(text), "_"))
}

// TokenizeString returns the words in the passed in string, split by non word characters including emojis
func TokenizeString(str string) []string {
	return wordTokenRegex.FindAllString(str, -1)
}

// TokenizeStringByChars returns the words in the passed in string, split by the chars in the given string
func TokenizeStringByChars(str string, chars string) []string {
	runes := []rune(chars)
	f := func(c rune) bool {
		for _, r := range runes {
			if c == r {
				return true
			}
		}
		return false
	}
	return strings.FieldsFunc(str, f)
}

// PrefixOverlap returns the number of prefix characters which s1 and s2 have in common
func PrefixOverlap(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)
	i := 0
	for ; i < len(r1) && i < len(r2) && r1[i] == r2[i]; i++ {
	}
	return i
}

// StringSlices returns the slices of s defined by pairs of indexes in indices
func StringSlices(s string, indices []int) []string {
	slices := make([]string, 0, len(indices)/2)
	for i := 0; i < len(indices); i += 2 {
		slices = append(slices, s[indices[i]:indices[i+1]])
	}
	return slices
}

// StringSliceContains determines whether the given slice of strings contains the given string
func StringSliceContains(slice []string, str string, caseSensitive bool) bool {
	for _, s := range slice {
		if (caseSensitive && s == str) || (!caseSensitive && strings.ToLower(s) == strings.ToLower(str)) {
			return true
		}
	}
	return false
}

// StringSliceEquals determines whether the given slices of strings are equal
func StringSliceEquals(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// StringSetKeys returns the keys of string set in lexical order
func StringSetKeys(m map[string]bool) []string {
	vals := make([]string, 0, len(m))
	for v := range m {
		vals = append(vals, v)
	}
	sort.Strings(vals)
	return vals
}

// Indent indents each non-empty line in the given string
func Indent(s string, prefix string) string {
	output := strings.Builder{}

	bol := true
	for _, c := range s {
		if bol && c != '\n' {
			output.WriteString(prefix)
		}
		output.WriteRune(c)
		bol = c == '\n'
	}
	return output.String()
}
