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

// The standard's position: the station is the line above the box, as in Pub 28
// §045. It is read the same way as the station below the box.
func TestPostalStationAboveBox(t *testing.T) {
	found := false
	for _, c := range candidates("OLD SAN JUAN STA\nPO BOX 1190\nSAN JUAN PR 00902-1190") {
		if c.Address.PostalStation == "OLD SAN JUAN STA" && c.Address.PrimaryNumber == "1190" {
			found = true
		}
	}

	if !found {
		t.Fatal("no candidate carries OLD SAN JUAN STA above the box as the postal station")
	}
}

// A station is strongly tied to a PO BOX street line. With an ordinary street
// line in the box's place there is no box reading at all, so no candidate can
// carry the station.
func TestPostalStationNeedsAPOBoxStreetLine(t *testing.T) {
	for _, c := range candidates("OLD SAN JUAN STA\n123 MAIN ST\nSAN JUAN PR 00902") {
		if c.Address.PostalStation != "" {
			t.Errorf("a station was read with a non-PO BOX street line: %q", c.Address.PostalStation)
		}
	}
}
