package streetsuffixes

import (
	"slices"
	"testing"
)

// This file is an internal test because the invariants it checks are about the
// table itself, not about what the package does with it. They are the two ways
// a hand maintained list of 200-odd entries goes wrong, and both had gone
// wrong: DAM did not list its own name, so a lookup of DAM failed, and DALE
// appeared under DAM, so a lookup of DALE returned the wrong suffix.

// TestTablePrimaryAndShortAreLookupKeys enforces the invariant documented on
// streetSuffixes: Alt is the only key the lookup maps are built from, so an
// entry that does not repeat its own Primary and Short into Alt cannot be
// found by them.
//
// The failure is silent in the direction that matters. NormalizeStreetSuffix
// reports an error for input it cannot find, and a caller reasonably reads
// that as "not a street suffix" rather than "the table forgot to index this
// one" — which is the reading pkg/address/parser/claim depends on, since a
// vocabulary that errors is saying the tokens are not its concern.
func TestTablePrimaryAndShortAreLookupKeys(t *testing.T) {
	for _, entry := range streetSuffixes {
		if !slices.Contains(entry.Alt, entry.Primary) {
			t.Errorf("%s/%s: Primary %q is not in Alt %v, so it cannot be looked up by its own name",
				entry.Primary, entry.Short, entry.Primary, entry.Alt)
		}
		if !slices.Contains(entry.Alt, entry.Short) {
			t.Errorf("%s/%s: Short %q is not in Alt %v, so it cannot be looked up by its own abbreviation",
				entry.Primary, entry.Short, entry.Short, entry.Alt)
		}
	}
}

// TestTableAltKeysHaveOneOwner enforces that no two entries claim the same Alt
// key.
//
// The lookup maps are built by iterating the table in order, so a duplicated
// key resolves to whichever entry appears later and the earlier one is
// unreachable through it. There is no error and no ambiguity reported — the
// caller gets a confident answer from the wrong entry, which is how DALE came
// to normalize to DAM while DALE had a perfectly good entry of its own three
// slots above.
func TestTableAltKeysHaveOneOwner(t *testing.T) {
	owner := map[string]string{}

	for _, entry := range streetSuffixes {
		for _, alt := range entry.Alt {
			if previous, seen := owner[alt]; seen {
				t.Errorf("alt key %q is claimed by both %q and %q; the later entry wins and the earlier is unreachable by that key",
					alt, previous, entry.Primary)
				continue
			}
			owner[alt] = entry.Primary
		}
	}
}
