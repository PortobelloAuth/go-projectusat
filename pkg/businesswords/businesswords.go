// Package businesswords reads a business word out of a business name.
//
// The rows are Publication 28 Appendix G, held once in
// github.com/poetic-systems/addresstables and read from there. What is here
// is the part that is about parsing: which spelling a token is, and
// NormalizeBusinessWordAbbreviation.
//
// Unlike pkg/streetsuffixes, the lookup map below is keyed by Primary, not
// Alt. The businesswords table does not give streetsuffixes' guarantees: 316
// Alt spellings are shared across more than one distinct Primary (e.g. ACCT
// is a listed common presentation under both ACCOUNT and ACCOUNTANT), so a
// map keyed by Alt would silently collide, and 71 rows have an official
// Short that isn't itself among that row's own Alt list. The one uniqueness
// guarantee the table does make is that Primary is unique across all rows
// (TestNoDuplicatePrimary in addresstables), so Primary is the only safe key
// available. This is a deliberate scope reduction versus streetsuffixes: a
// word is only recognized by its standard's own canonical primary spelling,
// not by every historical misspelling in Alt.
package businesswords

import (
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/poetic-systems/addresstables/businesswords"
)

type BusinessWord struct {
	Primary string
	Short   string
	Alt     []string
}

// businessWordPrimaryMap is Publication 28 Appendix G, keyed by Primary
// only. See the package doc for why Primary and not Alt is the lookup key.
var businessWordPrimaryMap = maps.Collect(func(yield func(string, BusinessWord) bool) {
	for w := range businesswords.All() {
		if !yield(w.Primary, BusinessWord{Primary: w.Primary, Short: w.Short, Alt: w.Alt}) {
			return
		}
	}
})

// punctuation matches everything a business word is not made of. Business
// word keys in this table are letters and spaces, so a digit or other
// punctuation surviving the strip is what makes a lookup fail.
var punctuation = regexp.MustCompile("[^a-zA-Z0-9 ]+")

// Info looks up a business word by its Primary spelling. The input is
// cleaned of punctuation and capitalized before the lookup, but is otherwise
// matched exactly: a word reachable only via Alt in the underlying table is
// not found here.
func Info(word string) (*BusinessWord, error) {
	clean := punctuation.ReplaceAllString(word, "")
	capitalized := strings.ToUpper(clean)

	info, ok := businessWordPrimaryMap[capitalized]
	if !ok {
		return nil, fmt.Errorf("Unrecognized business word")
	}

	return &BusinessWord{
		Primary: info.Primary,
		Short:   info.Short,
		Alt:     append([]string(nil), info.Alt...),
	}, nil
}

// NormalizeBusinessWordAbbreviation returns the standard abbreviation for a
// business word, or an error if the word is not a recognized Primary
// spelling.
func NormalizeBusinessWordAbbreviation(word string) (string, error) {
	info, err := Info(word)
	if err != nil {
		return "", err
	}

	return info.Short, nil
}
