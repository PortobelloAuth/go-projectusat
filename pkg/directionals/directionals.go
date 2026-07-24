package directionals

import (
	"fmt"
	"maps"
	"strings"
)

/*
North     N
East      E
South     S
West      W
Northeast NE
Southeast SE
Northwest NW
Southwest SW
*/

var directionMap = map[string]string{
	"NORTH":     "N",
	"EAST":      "E",
	"SOUTH":     "S",
	"WEST":      "W",
	"NORTHEAST": "NE",
	"SOUTHEAST": "SE",
	"NORTHWEST": "NW",
	"SOUTHWEST": "SW",
}

var directionShortMap = maps.Collect(func(yield func(string, string) bool) {
	for k, v := range directionMap {
		if !yield(v, k) {
			return
		}
	}
})

func AbbreviateDirectional(d string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(d)

	// look up the abbreviation
	abrev, ok := directionMap[capitalized]
	if ok {
		return abrev, nil
	}

	// See if it is an abbreviation
	_, ok = directionShortMap[capitalized]
	if ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized directional")
}

func NormalizeDirectional(d string) (string, error) {
	// capitalize
	capitalized := strings.ToUpper(d)

	// look up the primary (full) direction word
	full, ok := directionShortMap[capitalized]
	if ok {
		return full, nil
	}

	// See if it is the primary (full) direction word
	_, ok = directionMap[capitalized]
	if ok {
		return capitalized, nil
	}

	return "", fmt.Errorf("Unrecognized directional")
}

// Score returns how strongly token looks like a directional.
// 0 = not a directional; 100 = exact abbr/full match.
func Score(token string) (int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, nil
	}
	if _, err := AbbreviateDirectional(token); err != nil {
		return 0, nil
	}
	return 100, nil
}
