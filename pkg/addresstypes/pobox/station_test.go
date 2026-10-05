package pobox_test

import "testing"

// A postal station on the line between the box and the last line is carried as
// the station, not dropped. Without the station candidate the stranded line
// would leave the reading that drops it as the only one.
func TestPostalStationBetweenBoxAndLastLine(t *testing.T) {
	found := false
	for _, c := range candidates("PO BOX 1190\nOLD SAN JUAN STA\nSAN JUAN PR 00902-1190") {
		if c.Address.PostalStation == "OLD SAN JUAN STA" && c.Address.PrimaryNumber == "1190" {
			found = true
		}
	}

	if !found {
		t.Fatal("no candidate carries OLD SAN JUAN STA as the postal station")
	}
}

// A line that does not end in STA or STATION is not read as a station, so the
// same shape with a street name in that position carries no station.
func TestNoPostalStationWithoutDesignator(t *testing.T) {
	for _, c := range candidates("PO BOX 1190\nOLD SAN JUAN ROAD\nSAN JUAN PR 00902-1190") {
		if c.Address.PostalStation != "" {
			t.Errorf("read a postal station from a non-station line: %q", c.Address.PostalStation)
		}
	}
}
