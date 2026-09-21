package libpostalhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/libpostalhttp"
)

// TestParse checks Parse's mapping from a libpostal label/value response to
// an address.Address, against a canned response from an httptest server
// rather than a live libpostal service. The challenging live cases this test
// used to carry move with this package to its own repository (per Aaron's
// ask on #71); they stay in this file's git history.
func TestParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/parse" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]libpostalhttp.LibpostalAddressPart{
			{Label: "house_number", Value: "43"},
			{Label: "road", Value: "e 200 n"},
			{Label: "city", Value: "north salt lake"},
			{Label: "state", Value: "ut"},
		})
	}))
	defer srv.Close()

	p, err := libpostalhttp.NewService(srv.URL, time.Second)
	if err != nil {
		t.Fatalf("Unable to create libpostal http service: %v", err)
	}

	// Post-directional followed by a city with a directional prefix: the
	// same "43 E 200 N, NORTH SALT LAKE, UT" case the live test used to
	// exercise the two-word road split against.
	want := address.Address{
		PrimaryNumber:   "43",
		Predirectional:  "E",
		StreetName:      "200",
		Postdirectional: "N",
		City:            "NORTH SALT LAKE",
		Region:          "UT",
	}
	got, err := p.Parse("43 E 200 N, NORTH SALT LAKE, UT")
	if err != nil {
		t.Fatalf("Error parsing '43 E 200 N, NORTH SALT LAKE, UT': %s", err)
	}
	if !got.Equals(&want) {
		t.Errorf("Parse(...) = %s, want %s", *got, want)
	}
}

// TestHTTPExpand checks HTTPExpand decodes libpostal's /expand response, a
// plain JSON array of strings, against a canned response.
func TestHTTPExpand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/expand" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"43 east 200 north", "43 e 200 n"})
	}))
	defer srv.Close()

	p, err := libpostalhttp.NewService(srv.URL, time.Second)
	if err != nil {
		t.Fatalf("Unable to create libpostal http service: %v", err)
	}

	want := []string{"43 east 200 north", "43 e 200 n"}
	got, err := p.HTTPExpand("43 E 200 N")
	if err != nil {
		t.Fatalf("Error expanding '43 E 200 N': %s", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("HTTPExpand(...) = %v, want %v", got, want)
	}
}
