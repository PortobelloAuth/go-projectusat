package libpostalhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/directionals"
	"github.com/PortobelloAuth/go-projectusat/pkg/region"
	"github.com/PortobelloAuth/go-projectusat/pkg/secondaryunit"
	"github.com/PortobelloAuth/go-projectusat/pkg/streetsuffixes"
)

var boundaryWithSpace = regexp.MustCompile(`\b\s*`)
var whitespace = regexp.MustCompile(`\s+`)

type LibpostalAddressPart struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

const (
	// from https://github.com/openvenues/libpostal/blob/master/src/address_parser.h
	LIBPOSTAL_LABEL_HOUSE          = "house"
	LIBPOSTAL_LABEL_HOUSE_NUMBER   = "house_number"
	LIBPOSTAL_LABEL_PO_BOX         = "po_box"
	LIBPOSTAL_LABEL_BUILDING       = "building"
	LIBPOSTAL_LABEL_ENTRANCE       = "entrance"
	LIBPOSTAL_LABEL_STAIRCASE      = "staircase"
	LIBPOSTAL_LABEL_LEVEL          = "level"
	LIBPOSTAL_LABEL_UNIT           = "unit"
	LIBPOSTAL_LABEL_ROAD           = "road"
	LIBPOSTAL_LABEL_METRO_STATION  = "metro_station"
	LIBPOSTAL_LABEL_SUBURB         = "suburb"
	LIBPOSTAL_LABEL_CITY_DISTRICT  = "city_district"
	LIBPOSTAL_LABEL_CITY           = "city"
	LIBPOSTAL_LABEL_STATE_DISTRICT = "state_district"
	LIBPOSTAL_LABEL_ISLAND         = "island"
	LIBPOSTAL_LABEL_STATE          = "state"
	LIBPOSTAL_LABEL_POSTAL_CODE    = "postcode"
	LIBPOSTAL_LABEL_COUNTRY_REGION = "country_region"
	LIBPOSTAL_LABEL_COUNTRY        = "country"
	LIBPOSTAL_LABEL_WORLD_REGION   = "world_region"

	LIBPOSTAL_LABEL_WEBSITE   = "website"
	LIBPOSTAL_LABEL_TELEPHONE = "phone"
)

var specialCaseMap = map[string]string{
	"FARM TO MARKET": "FM",
}
var specialCases = slices.Collect(func(yield func(string) bool) {
	for k, v := range specialCaseMap {
		if !yield(k) {
			return
		}
		if !yield(v) {
			return
		}
	}
})
var specialCaseReplacer = strings.NewReplacer(specialCases...)

type LibpostalHttpService struct {
	BaseURL *url.URL
	Timeout time.Duration
}

// NOTE: libpostal does not implement the Project US@ specification. It does have a respected
// model for parsing addresses. Using the dockerized http API provided by pelias and
// Who's On First gives us a clean way to use the libpostal parser without requiring all users
// to use it as a shared object or C library.

func NewService(base string, timeout time.Duration) (*LibpostalHttpService, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse base URL for service")
	}

	return &LibpostalHttpService{
		BaseURL: u,
		Timeout: timeout,
	}, nil
}

func (l *LibpostalHttpService) call(endpoint string, address string) ([]byte, error) {
	client := &http.Client{Timeout: l.Timeout}

	p := endpoint
	if len(l.BaseURL.Path) > 0 {
		bp, _ := strings.CutPrefix(l.BaseURL.Path, "/")
		bp, _ = strings.CutSuffix(l.BaseURL.Path, "/")
		p = fmt.Sprintf("%s/%s", bp, endpoint)
	}

	u := &url.URL{
		Scheme: l.BaseURL.Scheme,
		Host:   l.BaseURL.Host,
		Path:   p,
	}

	q := u.Query()
	q.Set("address", address)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Request failed: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body: %w", err)
	}

	return body, nil
}

func (l *LibpostalHttpService) HTTPParse(address string) ([]LibpostalAddressPart, error) {
	body, err := l.call("parse", address)
	if err != nil {
		return nil, err
	}

	var parts []LibpostalAddressPart
	err = json.Unmarshal(body, &parts)
	if err != nil {
		return nil, fmt.Errorf("Unable to process parse endpoint response: %w", err)
	}

	return parts, nil
}

func (l *LibpostalHttpService) HTTPExpand(address string) ([]string, error) {
	body, err := l.call("expand", address)
	if err != nil {
		return nil, err
	}

	var expanded []string
	err = json.Unmarshal(body, &expanded)
	if err != nil {
		return nil, fmt.Errorf("Unable to process expand endpoint response: %w", err)
	}

	return expanded, nil
}

func (l *LibpostalHttpService) Parse(input string) (*address.Address, error) {
	fmt.Println("Using libpostalhttp.Parse()")

	// FIXME?: what do we do about mismatches in libpostal parsing like
	// the 1200 in "9062 farm to market 1200" being mis-parsed as a postal code?
	// Some of this may be driven by not providing a full address, but...
	uppercase := strings.ToUpper(input)
	replaced := specialCaseReplacer.Replace(uppercase)

	// FIXME: libpostal also completely fails to understand "Rural Route 91 Box A7",
	// where "Box A7" is technically the primary street number and RR 91 is the
	// standardized street name. libpostal parses "91" as the primary address, and
	// "Box A7" as a PO Box. It does understand "RR 3 BOX 98D", but is inconsistent
	// with leading alphabetical characters on the box number.

	parsed, err := l.HTTPParse(replaced)
	if err != nil {
		return nil, err
	}

	a := &address.Address{}
	// FIXME: this is an imperfect mapping!!
	for _, v := range parsed {
		part := strings.ToUpper(v.Value)
		switch v.Label {
		case LIBPOSTAL_LABEL_HOUSE:
			a.BusinessName = part
		case LIBPOSTAL_LABEL_HOUSE_NUMBER:
			a.PrimaryNumber = part
		case LIBPOSTAL_LABEL_UNIT:
			// split on word boundary, with or without spaces
			secparts := boundaryWithSpace.Split(part, 2)
			if len(secparts) < 1 {
				continue
			}
			// use normalization to determine if we need secondary designator
			info, err := secondaryunit.Info(secparts[0])
			if err != nil {
				return nil, fmt.Errorf("Error handling secondary unit value: %w", err)
			}
			a.SecondaryDesignator = info.Short
			if len(secparts) > 1 && info.Numbered {
				a.SecondaryNumber = secparts[1]
			}
		case LIBPOSTAL_LABEL_ROAD:
			stparts := whitespace.Split(part, -1)
			// find the street suffix
			// if there's a directional as the first (or first and second) parts
			// before the suffix and there is another part grouped with the suffix, treat it as the
			// predirectional
			// if there's a directional as the last (or last and second to last) parts
			// after the suffix and there is another part grouped with the suffix, treat it as the
			// postdirectional
			// FIXME: this doesn't necessarily handle puertorico addresses well
			found := -1
			numparts := len(stparts)
			if numparts < 2 {
				a.StreetName = part
				continue
			}
			aftersuffix := numparts
			beforesuffix := numparts
			suffix := ""
			for i, p := range stparts {
				info, _ := streetsuffixes.Info(p, false)
				if info != nil {
					found = i
					suffix = info.Short
				}
			}
			if found >= 0 {
				// We have a suffix somewhere
				aftersuffix = (numparts - 1) - found
				beforesuffix = found - 1
			}
			// check for two part predirectionals
			pre2part, _ := directionals.AbbreviateDirectional(strings.Join(stparts[0:2], ""))
			pre1part, _ := directionals.AbbreviateDirectional(strings.Join(stparts[0:1], ""))
			possiblestate, _ := region.Info(strings.Join(stparts[0:2], " "), false)
			post2part, _ := directionals.AbbreviateDirectional(strings.Join(stparts[numparts-2:], ""))
			post1part, _ := directionals.AbbreviateDirectional(strings.Join(stparts[numparts-1:], ""))

			fmt.Printf("pre2part: %s\npre1part: %s\npossiblestate: %v\n", pre2part, pre1part, possiblestate)
			fmt.Printf("post2part:%s \npost1part: %s\n", post2part, post1part)
			fmt.Printf("found:%d \nsuffix: %s\n", found, suffix)
			fmt.Printf("numparts: %d\nbeforesuffix: %d\naftersuffix: %d\n", numparts, beforesuffix, aftersuffix)

			nondirparts := stparts[0:]
			// FIXME?: the before/after suffix logic may not be correct when no suffix is found
			if len(pre2part) > 0 && (beforesuffix > 2 || (beforesuffix == 2 && aftersuffix >= max(len(post2part), len(post1part)))) {
				// two-part predirectional
				a.Predirectional = pre2part
				nondirparts = nondirparts[2:]
				found -= 2
				if len(post2part) > 0 {
					// and 2-part postdirectional
					a.Postdirectional = post2part
					nondirparts = nondirparts[0 : len(nondirparts)-2]
				} else if len(post1part) > 0 {
					// and 1-part postdirectional
					a.Postdirectional = post1part
					nondirparts = nondirparts[0 : len(nondirparts)-1]
				} else {
					// no post directional
				}
			} else if len(pre1part) > 0 && possiblestate == nil && (beforesuffix > 1 || (beforesuffix == 1 && aftersuffix >= max(len(post2part), len(post1part)))) {
				// 1 part predirectional
				a.Predirectional = pre1part
				nondirparts = nondirparts[1:]
				found -= 1
				if len(post2part) > 0 {
					// and 2-part postdirectional
					a.Postdirectional = post2part
					nondirparts = nondirparts[0 : len(nondirparts)-2]
				} else if len(post1part) > 0 {
					// and 1-part postdirectional
					a.Postdirectional = post1part
					nondirparts = nondirparts[0 : len(nondirparts)-1]
				} else {
					// no post directional
				}
			} else {
				// no predirectional
				if len(post2part) > 0 {
					// and 2-part postdirectional
					a.Postdirectional = post2part
					nondirparts = nondirparts[0 : len(nondirparts)-2]
				} else if len(post1part) > 0 {
					// and 1-part postdirectional
					a.Postdirectional = post1part
					nondirparts = nondirparts[0 : len(nondirparts)-1]
				} else {
					// no post directional
				}
			}

			fmt.Printf("nondirparts: %v\n", nondirparts)
			if found >= 0 && len(nondirparts) > 1 {
				nondirparts[found] = suffix
			}

			a.StreetName = strings.Join(nondirparts, " ")
		case LIBPOSTAL_LABEL_CITY:
			a.City = part
		case LIBPOSTAL_LABEL_STATE:
			a.Region = part
		case LIBPOSTAL_LABEL_POSTAL_CODE:
			a.Postal = part
		case LIBPOSTAL_LABEL_COUNTRY:
			a.Country = part

		// TODO?: handle POBOX, LEVEL, and other options
		case LIBPOSTAL_LABEL_PO_BOX:
			a.StreetName = part
		case LIBPOSTAL_LABEL_BUILDING:
			a.SecondaryDesignator = part
		case LIBPOSTAL_LABEL_ENTRANCE:
			a.SecondaryDesignator = part
		case LIBPOSTAL_LABEL_STAIRCASE:
			a.SecondaryDesignator = part
		case LIBPOSTAL_LABEL_LEVEL:
			a.SecondaryDesignator = part
		// case LIBPOSTAL_LABEL_METRO_STATION:
		// case LIBPOSTAL_LABEL_SUBURB:
		// case LIBPOSTAL_LABEL_CITY_DISTRICT:
		// case LIBPOSTAL_LABEL_STATE_DISTRICT:
		// case LIBPOSTAL_LABEL_ISLAND:
		// case LIBPOSTAL_LABEL_COUNTRY_REGION:
		// case LIBPOSTAL_LABEL_WORLD_REGION:
		default:
		}
	}

	return a, nil
}
