package parsertest

// https://github.com/GreenBuildingRegistry/usaddress-scourgify/commit/cf66bdc42aebc57d48fe41e32e108fdcde347e3c (MIT License)
// lists a couple of addresses on it's README that give it trouble. They are used here to test whether using zipcity
// helps us avoid such difficulties.
var ScourgifyCases = []Case{
	{
		Source: "scourgify",
		Note:   "multi-word street and city names can be difficult to properly distinguish",
		Input:  "12790 LA TORTOLA, SAN DIEGO, CA 92129",
		Want:   "12790 LA TORTOLA\nSAN DIEGO CA 92129",
	},
	{
		Source: "scourgify",
		Note:   "multi-word street and city names can be difficult to properly distinguish",
		Input:  "12790 LA TORTOLA SAN DIEGO CA 92129",
		Want:   "12790 LA TORTOLA\nSAN DIEGO CA 92129",
	},
	{
		Source: "scourgify",
		Note:   "VIA is both and abbreviation and a word often used in street names; VISTA is both the name of a city in California and a known street suffix",
		Input:  "1509 Via Christina, Vista, CA 92084",
		Want:   "1509 VIA CHRISTINA\nVISTA CA 92084",
	},
	{
		Source: "scourgify",
		Note:   "VIA is both and abbreviation and a word often used in street names; VISTA is both the name of a city in California and a known street suffix",
		Input:  "1509 Via Christina Vista CA 92084",
		Want:   "1509 VIA CHRISTINA\nVISTA CA 92084",
	},
}
