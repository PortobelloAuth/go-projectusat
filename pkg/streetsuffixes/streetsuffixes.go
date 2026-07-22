package streetsuffixes

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hbollon/go-edlib"
)

type StreetSuffix struct {
	Primary string
	Short   string
	Alt     []string
}

var streetSuffixes = []StreetSuffix{
	{
		Primary: "ALLEY",
		Short:   "ALY",
		Alt: []string{
			"ALLEE", "ALLEY", "ALLY", "ALY",
		},
	},
	{
		Primary: "ANEX",
		Short:   "ANX",
		Alt: []string{
			"ANEX", "ANNEX", "ANNX", "ANX",
		},
	},
	{
		Primary: "ARCADE",
		Short:   "ARC",
		Alt: []string{
			"ARC", "ARCADE",
		},
	},
	{
		Primary: "AVENUE",
		Short:   "AVE",
		Alt: []string{
			"AV", "AVE", "AVEN", "AVENU", "AVENUE", "AVN", "AVNUE",
		},
	},
	{
		Primary: "BAYOU",
		Short:   "BYU",
		Alt: []string{
			"BAYOO", "BAYOU", "BYU",
		},
	},
	{
		Primary: "BEACH",
		Short:   "BCH",
		Alt: []string{
			"BCH", "BEACH",
		},
	},
	{
		Primary: "BEND",
		Short:   "BND",
		Alt: []string{
			"BEND", "BND",
		},
	},
	{
		Primary: "BLUFF",
		Short:   "BLF",
		Alt: []string{
			"BLF", "BLUF", "BLUFF",
		},
	},
	{
		Primary: "BLUFFS",
		Short:   "BLFS",
		Alt: []string{
			"BLFS", "BLUFFS",
		},
	},
	{
		Primary: "BOTTOM",
		Short:   "BTM",
		Alt: []string{
			"BOT", "BTM", "BOTTM", "BOTTOM",
		},
	},
	{
		Primary: "BOULEVARD",
		Short:   "BLVD",
		Alt: []string{
			"BLVD", "BOUL", "BOULEVARD", "BOULV",
		},
	},
	{
		Primary: "BRANCH",
		Short:   "BR",
		Alt: []string{
			"BR", "BRNCH", "BRANCH",
		},
	},
	{
		Primary: "BRIDGE",
		Short:   "BRG",
		Alt: []string{
			"BRG", "BRDGE", "BRIDGE",
		},
	},
	{
		Primary: "BROOK",
		Short:   "BRK",
		Alt: []string{
			"BRK", "BROOK",
		},
	},
	{
		Primary: "BROOKS",
		Short:   "BRKS",
		Alt: []string{
			"BRKS", "BROOKS",
		},
	},
	{
		Primary: "BURG",
		Short:   "BG",
		Alt: []string{
			"BG", "BURG",
		},
	},
	{
		Primary: "BURGS",
		Short:   "BGS",
		Alt: []string{
			"BGS", "BURGS",
		},
	},
	{
		Primary: "BYPASS",
		Short:   "BYP",
		Alt: []string{
			"BYP", "BYPA", "BYPAS", "BYPASS", "BYPS",
		},
	},
	{
		Primary: "CAMP",
		Short:   "CP",
		Alt: []string{
			"CP", "CMP", "CAMP",
		},
	},
	{
		Primary: "CANYON",
		Short:   "CYN",
		Alt: []string{
			"CYN", "CANYN", "CNYN", "CANYON",
		},
	},
	{
		Primary: "CAPE",
		Short:   "CPE",
		Alt: []string{
			"CPE", "CAPE",
		},
	},
	{
		Primary: "CAUSEWAY",
		Short:   "CSWY",
		Alt: []string{
			"CSWY", "CAUSWA", "CAUSEWAY",
		},
	},
	{
		Primary: "CENTER",
		Short:   "CTR",
		Alt: []string{
			"CEN", "CENT", "CENTER", "CENTR", "CENTRE", "CNTER", "CNTR", "CTR",
		},
	},
	{
		Primary: "CENTERS",
		Short:   "CTRS",
		Alt: []string{
			"CTRS", "CENTERS",
		},
	},
	{
		Primary: "CIRCLE",
		Short:   "CIR",
		Alt: []string{
			"CIR", "CIRC", "CIRCL", "CIRCLE", "CRCL", "CRCLE",
		},
	},
	{
		Primary: "CIRCLES",
		Short:   "CIRS",
		Alt: []string{
			"CIRS", "CIRCLES",
		},
	},
	{
		Primary: "CLIFF",
		Short:   "CLF",
		Alt: []string{
			"CLF", "CLIFF",
		},
	},
	{
		Primary: "CLIFFS",
		Short:   "CLFS",
		Alt: []string{
			"CLFS", "CLIFFS",
		},
	},
	{
		Primary: "CLUB",
		Short:   "CLB",
		Alt: []string{
			"CLB", "CLUB",
		},
	},
	{
		Primary: "COMMON",
		Short:   "CMN",
		Alt: []string{
			"CMN", "COMMON",
		},
	},
	{
		Primary: "COMMONS",
		Short:   "CMNS",
		Alt: []string{
			"CMNS", "COMMONS",
		},
	},
	{
		Primary: "CORNER",
		Short:   "COR",
		Alt: []string{
			"COR", "CORNER",
		},
	},
	{
		Primary: "CORNERS",
		Short:   "CORS",
		Alt: []string{
			"CORS", "CORNERS",
		},
	},
	{
		Primary: "COURSE",
		Short:   "CRSE",
		Alt: []string{
			"CRSE", "COURSE",
		},
	},
	{
		Primary: "COURT",
		Short:   "CT",
		Alt: []string{
			"CT", "COURT",
		},
	},
	{
		Primary: "COURTS",
		Short:   "CTS",
		Alt: []string{
			"CTS", "COURTS",
		},
	},
	{
		Primary: "COVE",
		Short:   "CV",
		Alt: []string{
			"CV", "COVE",
		},
	},
	{
		Primary: "COVES",
		Short:   "CVS",
		Alt: []string{
			"CVS", "COVES",
		},
	},
	{
		Primary: "CREEK",
		Short:   "CRK",
		Alt: []string{
			"CRK", "CREEK",
		},
	},
	{
		Primary: "CRESCENT",
		Short:   "CRES",
		Alt: []string{
			"CRES", "CRSENT", "CRSNT", "CRESCENT",
		},
	},
	{
		Primary: "CREST",
		Short:   "CRST",
		Alt: []string{
			"CRST", "CREST",
		},
	},
	{
		Primary: "CROSSING",
		Short:   "XING",
		Alt: []string{
			"CRSSNG", "CROSSING", "XING",
		},
	},
	{
		Primary: "CROSSROAD",
		Short:   "XRD",
		Alt: []string{
			"CROSSROAD", "XRD",
		},
	},
	{
		Primary: "CROSSROADS",
		Short:   "XRDS",
		Alt: []string{
			"CROSSROADS", "XRDS",
		},
	},
	{
		Primary: "CURVE",
		Short:   "CURV",
		Alt: []string{
			"CURV", "CURVE",
		},
	},
	{
		Primary: "DALE",
		Short:   "DL",
		Alt: []string{
			"DL", "DALE",
		},
	},
	{
		Primary: "DAM",
		Short:   "DM",
		Alt: []string{
			"DM", "DALE",
		},
	},
	{
		Primary: "DIVIDE",
		Short:   "DV",
		Alt: []string{
			"DV", "DIV", "DVD", "DIVIDE",
		},
	},
	{
		Primary: "DRIVE",
		Short:   "DR",
		Alt: []string{
			"DR", "DRIV", "DRIVE", "DRV",
		},
	},
	{
		Primary: "DRIVES",
		Short:   "DRS",
		Alt: []string{
			"DRS", "DRIVES",
		},
	},
	{
		Primary: "ESTATE",
		Short:   "EST",
		Alt: []string{
			"EST", "ESTATE",
		},
	},
	{
		Primary: "ESTATES",
		Short:   "ESTS",
		Alt: []string{
			"ESTS", "ESTATES",
		},
	},
	{
		Primary: "EXPRESSWAY",
		Short:   "EXPY",
		Alt: []string{
			"EXP", "EXPY", "EXPR", "EXPRESS", "EXPW", "EXPRESSWAY",
		},
	},
	{
		Primary: "EXTENSION",
		Short:   "EXT",
		Alt: []string{
			"EXT", "EXTN", "EXTNSN", "EXTENSION",
		},
	},
	{
		Primary: "EXTENSIONS",
		Short:   "EXTS",
		Alt: []string{
			"EXTS", "EXTENSIONS",
		},
	},
	{
		Primary: "FALL",
		Short:   "FALL",
		Alt: []string{
			"FALL",
		},
	},
	{
		Primary: "FALLS",
		Short:   "FLS",
		Alt: []string{
			"FLS", "FALLS",
		},
	},
	{
		Primary: "FERRY",
		Short:   "FRY",
		Alt: []string{
			"FRY", "FRRY", "FERRY",
		},
	},
	{
		Primary: "FIELD",
		Short:   "FLD",
		Alt: []string{
			"FLD", "FIELD",
		},
	},
	{
		Primary: "FIELDS",
		Short:   "FLDS",
		Alt: []string{
			"FLDS", "FIELDS",
		},
	},
	{
		Primary: "FLAT",
		Short:   "FLT",
		Alt: []string{
			"FLT", "FLAT",
		},
	},
	{
		Primary: "FLATS",
		Short:   "FLTS",
		Alt: []string{
			"FLTS", "FLATS",
		},
	},
	{
		Primary: "FORD",
		Short:   "FRD",
		Alt: []string{
			"FRD", "FORD",
		},
	},
	{
		Primary: "FORDS",
		Short:   "FRDS",
		Alt: []string{
			"FRDS", "FORDS",
		},
	},
	{
		Primary: "FOREST",
		Short:   "FRST",
		Alt: []string{
			"FRST", "FOREST", "FORESTS",
		},
	},
	{
		Primary: "FORGE",
		Short:   "FRG",
		Alt: []string{
			"FRG", "FORG", "FORGE",
		},
	},
	{
		Primary: "FORGES",
		Short:   "FRGS",
		Alt: []string{
			"FRGS", "FORGES",
		},
	},
	{
		Primary: "FORK",
		Short:   "FRK",
		Alt: []string{
			"FRK", "FORK",
		},
	},
	{
		Primary: "FORKS",
		Short:   "FRKS",
		Alt: []string{
			"FRKS", "FORKS",
		},
	},
	{
		Primary: "FORK",
		Short:   "FRK",
		Alt: []string{
			"FRK", "FORK",
		},
	},
	{
		Primary: "FORKS",
		Short:   "FRKS",
		Alt: []string{
			"FRKS", "FORKS",
		},
	},
	{
		Primary: "FORT",
		Short:   "FT",
		Alt: []string{
			"FT", "FORT", "FRT",
		},
	},
	{
		Primary: "FREEWAY",
		Short:   "FWY",
		Alt: []string{
			"FWY", "FRWY", "FRWAY", "FREEWY", "FREEWAY",
		},
	},
	{
		Primary: "GARDEN",
		Short:   "GDN",
		Alt: []string{
			"GDN", "GRDN", "GARDN", "GRDEN", "GARDEN",
		},
	},
	{
		Primary: "GARDENS",
		Short:   "GDNS",
		Alt: []string{
			"GDNS", "GRDNS", "GARDENS",
		},
	},
	{
		Primary: "GATEWAY",
		Short:   "GTWY",
		Alt: []string{
			"GTWY", "GTWAY", "GATWY", "GATEWY", "GATEWAY",
		},
	},
	{
		Primary: "GLEN",
		Short:   "GLN",
		Alt: []string{
			"GLN", "GLEN",
		},
	},
	{
		Primary: "GLENS",
		Short:   "GLNS",
		Alt: []string{
			"GLNS", "GLENS",
		},
	},
	{
		Primary: "GREEN",
		Short:   "GRN",
		Alt: []string{
			"GRN", "GREEN",
		},
	},
	{
		Primary: "GREENS",
		Short:   "GRNS",
		Alt: []string{
			"GRNS", "GREENS",
		},
	},
	{
		Primary: "GROVE",
		Short:   "GRV",
		Alt: []string{
			"GRV", "GROV", "GROVE",
		},
	},
	{
		Primary: "GROVES",
		Short:   "GRVS",
		Alt: []string{
			"GRVS", "GROVES",
		},
	},
	{
		Primary: "HARBOR",
		Short:   "HBR",
		Alt: []string{
			"HBR", "HARB", "HRBOR", "HARBR", "HARBOR",
		},
	},
	{
		Primary: "HARBORS",
		Short:   "HBRS",
		Alt: []string{
			"HBRS", "HARBORS",
		},
	},
	{
		Primary: "HAVEN",
		Short:   "HVN",
		Alt: []string{
			"HVN", "HAVEN",
		},
	},
	{
		Primary: "HEIGHTS",
		Short:   "HTS",
		Alt: []string{
			"HT", "HTS", "HEIGHTS",
		},
	},
	{
		Primary: "HIGHWAY",
		Short:   "HWY",
		Alt: []string{
			"HWY", "HWAY", "HIWY", "HIWAY", "HIGHWY", "HIGHWAY",
		},
	},
	{
		Primary: "HILL",
		Short:   "HL",
		Alt: []string{
			"HL", "HILL",
		},
	},
	{
		Primary: "HILLS",
		Short:   "HLS",
		Alt: []string{
			"HLS", "HILLS",
		},
	},
	{
		Primary: "HOLLOW",
		Short:   "HOLW",
		Alt: []string{
			"HLLW", "HOLLOW", "HOLLOWS", "HOLW", "HOLWS",
		},
	},
	{
		Primary: "INLET",
		Short:   "INLT",
		Alt: []string{
			"INLT", "INLET",
		},
	},
	{
		Primary: "ISLAND",
		Short:   "IS",
		Alt: []string{
			"IS", "ISLND", "ISLAND",
		},
	},
	{
		Primary: "ISLANDS",
		Short:   "ISS",
		Alt: []string{
			"ISS", "ISLNDS", "ISLANDS",
		},
	},
	{
		Primary: "ISLE",
		Short:   "ISLE",
		Alt: []string{
			"ISLE", "ISLES",
		},
	},
	{
		Primary: "JUNCTION",
		Short:   "JCT",
		Alt: []string{
			"JCT", "JCTN", "JCTION", "JUNCTION", "JUNCTN", "JUNCTON",
		},
	},
	{
		Primary: "JUNCTIONS",
		Short:   "JCTS",
		Alt: []string{
			"JCTS", "JCTNS", "JUNCTIONS",
		},
	},
	{
		Primary: "KEY",
		Short:   "KY",
		Alt: []string{
			"KY", "KEY",
		},
	},
	{
		Primary: "KEYS",
		Short:   "KYS",
		Alt: []string{
			"KYS", "KEYS",
		},
	},
	{
		Primary: "KNOLL",
		Short:   "KNL",
		Alt: []string{
			"KNL", "KNOL", "KNOLL",
		},
	},
	{
		Primary: "KNOLLS",
		Short:   "KNLS",
		Alt: []string{
			"KNLS", "KNOLLS",
		},
	},
	{
		Primary: "LAKE",
		Short:   "LK",
		Alt: []string{
			"LK", "LAKE",
		},
	},
	{
		Primary: "LAKES",
		Short:   "LKS",
		Alt: []string{
			"LKS", "LAKES",
		},
	},
	{
		Primary: "LAND",
		Short:   "LAND",
		Alt: []string{
			"LAND",
		},
	},
	{
		Primary: "LANDING",
		Short:   "LNDG",
		Alt: []string{
			"LNDG", "LNDNG", "LANDING",
		},
	},
	{
		Primary: "LANE",
		Short:   "LN",
		Alt: []string{
			"LN", "LANE",
		},
	},
	{
		Primary: "LIGHT",
		Short:   "LGT",
		Alt: []string{
			"LGT", "LIGHT",
		},
	},
	{
		Primary: "LIGHTS",
		Short:   "LGTS",
		Alt: []string{
			"LGTS", "LIGHTS",
		},
	},
	{
		Primary: "LOAF",
		Short:   "LF",
		Alt: []string{
			"LF", "LOAF",
		},
	},
	{
		Primary: "LOCK",
		Short:   "LCK",
		Alt: []string{
			"LCK", "LOCK",
		},
	},
	{
		Primary: "LOCKS",
		Short:   "LCKS",
		Alt: []string{
			"LOCKS", "LCKS",
		},
	},
	{
		Primary: "LODGE",
		Short:   "LDG",
		Alt: []string{
			"LDG", "LDGE", "LODG", "LODGE",
		},
	},
	{
		Primary: "LOOP",
		Short:   "LOOP",
		Alt: []string{
			"LOOP", "LOOPS",
		},
	},
	{
		Primary: "MALL",
		Short:   "MALL",
		Alt: []string{
			"MALL",
		},
	},
	{
		Primary: "MANOR",
		Short:   "MNR",
		Alt: []string{
			"MNR", "MANOR",
		},
	},
	{
		Primary: "MANORS",
		Short:   "MNRS",
		Alt: []string{
			"MNRS", "MANORS",
		},
	},
	{
		Primary: "MEADOW",
		Short:   "MDW",
		Alt: []string{
			"MDW", "MEADOW",
		},
	},
	{
		Primary: "MEADOWS",
		Short:   "MDWS",
		Alt: []string{
			"MDWS", "MEDOWS", "MEADOWS",
		},
	},
	{
		Primary: "MEWS",
		Short:   "MEWS",
		Alt: []string{
			"MEWS",
		},
	},
	{
		Primary: "MILL",
		Short:   "ML",
		Alt: []string{
			"ML", "MILL",
		},
	},
	{
		Primary: "MILLS",
		Short:   "MLS",
		Alt: []string{
			"MLS", "MILLS",
		},
	},
	{
		Primary: "MISSION",
		Short:   "MSN",
		Alt: []string{
			"MSN", "MISSN", "MSSN", "MISSION",
		},
	},
	{
		Primary: "MOTORWAY",
		Short:   "MTWY",
		Alt: []string{
			"MTWY", "MOTORWAY",
		},
	},
	{
		Primary: "MOUNT",
		Short:   "MT",
		Alt: []string{
			"MT", "MNT", "MOUNT",
		},
	},
	{
		Primary: "MOUNTAIN",
		Short:   "MTN",
		Alt: []string{
			"MTN", "MNTN", "MTIN", "MNTAIN", "MOUNTIN", "MOUNTAIN",
		},
	},
	{
		Primary: "MOUNTAINS",
		Short:   "MTNS",
		Alt: []string{
			"MTNS", "MNTNS", "MOUNTAINS",
		},
	},
	{
		Primary: "NECK",
		Short:   "NCK",
		Alt: []string{
			"NCK", "NECK",
		},
	},
	{
		Primary: "ORCHARD",
		Short:   "ORCH",
		Alt: []string{
			"ORCH", "ORCHRD", "ORCHARD",
		},
	},
	{
		Primary: "OVAL",
		Short:   "OVAL",
		Alt: []string{
			"OVL", "OVAL",
		},
	},
	{
		Primary: "OVERPASS",
		Short:   "OPAS",
		Alt: []string{
			"OPAS", "OVERPAS",
		},
	},
	{
		Primary: "PARK",
		Short:   "PARK",
		Alt: []string{
			"PRK", "PARK",
		},
	},
	{
		Primary: "PARKS",
		Short:   "PARKS",
		Alt: []string{
			"PARKS",
		},
	},
	{
		Primary: "PARKWAY",
		Short:   "PKWY",
		Alt: []string{
			"PKY", "PKWY", "PARKWY", "PKWAY", "PARKWAY",
		},
	},
	{
		Primary: "PARKWAYS",
		Short:   "PKWYS",
		Alt: []string{
			"PKWYS", "PARKWAYS",
		},
	},
	{
		Primary: "PASS",
		Short:   "PASS",
		Alt: []string{
			"PASS",
		},
	},
	{
		Primary: "PASSAGE",
		Short:   "PSGE",
		Alt: []string{
			"PSGE", "PASSAGE",
		},
	},
	{
		Primary: "PATH",
		Short:   "PATH",
		Alt: []string{
			"PATH", "PATHS",
		},
	},
	{
		Primary: "PIKE",
		Short:   "PIKE",
		Alt: []string{
			"PIKE", "PIKES",
		},
	},
	{
		Primary: "PINE",
		Short:   "PNE",
		Alt: []string{
			"PNE", "PINE",
		},
	},
	{
		Primary: "PINES",
		Short:   "PNES",
		Alt: []string{
			"PNES", "PINES",
		},
	},
	{
		Primary: "PLACE",
		Short:   "PL",
		Alt: []string{
			"PL", "PLACE",
		},
	},
	{
		Primary: "PLAIN",
		Short:   "PLN",
		Alt: []string{
			"PLN", "PLAIN",
		},
	},
	{
		Primary: "PLAINS",
		Short:   "PLNS",
		Alt: []string{
			"PLNS", "PLAINS",
		},
	},
	{
		Primary: "PLAZA",
		Short:   "PLZ",
		Alt: []string{
			"PLZ", "PLZA", "PLAZA",
		},
	},
	{
		Primary: "POINT",
		Short:   "PT",
		Alt: []string{
			"PT", "POINT",
		},
	},
	{
		Primary: "POINTS",
		Short:   "PTS",
		Alt: []string{
			"PTS", "POINTS",
		},
	},
	{
		Primary: "PORT",
		Short:   "PRT",
		Alt: []string{
			"PRT", "PORT",
		},
	},
	{
		Primary: "PORTS",
		Short:   "PRTS",
		Alt: []string{
			"PRTS", "PORTS",
		},
	},
	{
		Primary: "PRAIRE",
		Short:   "PR",
		Alt: []string{
			"PR", "PRR", "PRAIRE",
		},
	},
	{
		Primary: "RADIAL",
		Short:   "RADL",
		Alt: []string{
			"RAD", "RADL", "RADIAL", "RADIEL",
		},
	},
	{
		Primary: "RAMP",
		Short:   "RAMP",
		Alt: []string{
			"RAMP",
		},
	},
	{
		Primary: "RANCH",
		Short:   "RNCH",
		Alt: []string{
			"RNCH", "RNCHS", "RANCH", "RANCHES",
		},
	},
	{
		Primary: "RAPID",
		Short:   "RPD",
		Alt: []string{
			"RPD", "RAPID",
		},
	},
	{
		Primary: "RAPIDS",
		Short:   "RPDS",
		Alt: []string{
			"RPDS", "RAPIDS",
		},
	},
	{
		Primary: "REST",
		Short:   "RST",
		Alt: []string{
			"RST", "REST",
		},
	},
	{
		Primary: "RIDGE",
		Short:   "RDG",
		Alt: []string{
			"RDG", "RDGE", "RIDGE",
		},
	},
	{
		Primary: "RIDGES",
		Short:   "RDGS",
		Alt: []string{
			"RDGS", "RIDGES",
		},
	},
	{
		Primary: "RIVER",
		Short:   "RIV",
		Alt: []string{
			"RIV", "RVR", "RIVR", "RIVER",
		},
	},
	{
		Primary: "ROAD",
		Short:   "RD",
		Alt: []string{
			"RD", "ROAD",
		},
	},
	{
		Primary: "ROADS",
		Short:   "RDS",
		Alt: []string{
			"RDS", "ROADS",
		},
	},
	{
		Primary: "ROUTE",
		Short:   "RTE",
		Alt: []string{
			"RTE", "ROUTE",
		},
	},
	{
		Primary: "NECK",
		Short:   "NCK",
		Alt: []string{
			"NCK", "NECK",
		},
	},
	{
		Primary: "ROW",
		Short:   "ROW",
		Alt: []string{
			"ROW",
		},
	},
	{
		Primary: "RUE",
		Short:   "RUE",
		Alt: []string{
			"RUE",
		},
	},
	{
		Primary: "RUN",
		Short:   "RUN",
		Alt: []string{
			"RUN",
		},
	},
	{
		Primary: "SHOAL",
		Short:   "SHL",
		Alt: []string{
			"SHL", "SHOAL",
		},
	},
	{
		Primary: "SHOALS",
		Short:   "SHLS",
		Alt: []string{
			"SHLS", "SHOALS",
		},
	},
	{
		Primary: "SHORE",
		Short:   "SHR",
		Alt: []string{
			"SHR", "SHORE", "SHOAR",
		},
	},
	{
		Primary: "SHORES",
		Short:   "SHRS",
		Alt: []string{
			"SHRS", "SHORES", "SHOARS",
		},
	},
	{
		Primary: "SKYWAY",
		Short:   "SKWY",
		Alt: []string{
			"SKWY", "SKYWAY",
		},
	},
	{
		Primary: "SPRING",
		Short:   "SPG",
		Alt: []string{
			"SPG", "SPNG", "SPRNG", "SPRING",
		},
	},
	{
		Primary: "SPRINGS",
		Short:   "SPGS",
		Alt: []string{
			"SPGS", "SPNGS", "SPRNGS", "SPRINGS",
		},
	},
	{
		Primary: "SPUR",
		Short:   "SPUR",
		Alt: []string{
			"SPUR",
			"SPURS", // see note
		},
	},
	// {
	// 	// This is an example of a plural being abbreviated to a
	// 	// singlular - of which the standard has several - that
	// 	// make the standard "unstable"; you cannot know whether
	// 	// the abbreviation refers to the plural or the singular.
	// 	// Futher, you cannot recover the Primary value reliably
	// 	// from the abbreviation.
	// 	Primary: "SPURS",
	// 	Short: "SPUR",
	// 	Alt: []string{
	// 		"SPURS",
	// 	},
	// },
	{
		Primary: "SQUARE",
		Short:   "SQ",
		Alt: []string{
			"SQ", "SQR", "SQRE", "SQU", "SQUARE",
		},
	},
	{
		Primary: "SQUARES",
		Short:   "SQS",
		Alt: []string{
			"SQS", "SQRS", "SQUARES",
		},
	},
	{
		Primary: "STATION",
		Short:   "STA",
		Alt: []string{
			"STA", "STN", "STATN", "STATION",
		},
	},
	{
		Primary: "STRAVENUE",
		Short:   "STRA",
		Alt: []string{
			"STRA", "STRAV", "STRAVN", "STRVN", "STRVNUE", "STRAVEN", "STRAVENUE",
		},
	},
	{
		Primary: "STREAM",
		Short:   "STRM",
		Alt: []string{
			"STRM", "STREME", "STREAM",
		},
	},
	{
		Primary: "STREET",
		Short:   "ST",
		Alt: []string{
			"ST", "STR", "STRT", "STREET",
		},
	},
	{
		Primary: "STREETS",
		Short:   "STS",
		Alt: []string{
			"STS", "STREETS",
		},
	},
	{
		Primary: "SUMMIT",
		Short:   "SMT",
		Alt: []string{
			"SMT", "SUMIT", "SUMITT", "SUMMIT",
		},
	},
	{
		Primary: "TERRACE",
		Short:   "TERR",
		Alt: []string{
			"TER", "TERR", "TERRACE",
		},
	},
	{
		Primary: "THROUGHWAY",
		Short:   "TRWY",
		Alt: []string{
			"TRWY", "THROUGHWAY",
		},
	},
	{
		Primary: "TRACE",
		Short:   "TRCE",
		Alt: []string{
			"TRCE", "TRACE", "TRACES",
		},
	},
	{
		Primary: "TRACK",
		Short:   "TRAK",
		Alt: []string{
			"TRK", "TRKS", "TRAK", "TRACK", "TRACKS",
		},
	},
	{
		Primary: "TRAFFICWAY",
		Short:   "TRFY",
		Alt: []string{
			"TRFY", "TRAFFICWAY",
		},
	},
	{
		Primary: "TRAIL",
		Short:   "TRL",
		Alt: []string{
			"TRL", "TRLS", "TRAIL", "TRAILS",
		},
	},
	{
		Primary: "TRAILER",
		Short:   "TRLR",
		Alt: []string{
			"TRLR", "TRLRS", "TRAILER",
		},
	},
	{
		Primary: "TUNNEL",
		Short:   "TUNL",
		Alt: []string{
			"TUNL", "TUNNL", "TUNLS", "TUNEL", "TUNNEL", "TUNNELS",
		},
	},
	{
		Primary: "TURNPIKE",
		Short:   "TPKE",
		Alt: []string{
			"TPKE", "TRNPK", "TURNPK", "TURNPIKE",
		},
	},
	{
		Primary: "UNDERPASS",
		Short:   "UPAS",
		Alt: []string{
			"UPAS", "UNDERPASS",
		},
	},
	{
		Primary: "UNION",
		Short:   "UN",
		Alt: []string{
			"UN", "UNION",
		},
	},
	{
		Primary: "UNIONS",
		Short:   "UNS",
		Alt: []string{
			"UNS", "UNIONS",
		},
	},
	{
		Primary: "SQUARE",
		Short:   "SQ",
		Alt: []string{
			"SQ", "SQR", "SQRE", "SQU", "SQUARE",
		},
	},
	{
		Primary: "VALLEY",
		Short:   "VLY",
		Alt: []string{
			"VLY", "VLLY", "VALLY", "VALLEY",
		},
	},
	{
		Primary: "VALLEYS",
		Short:   "VLYS",
		Alt: []string{
			"VLYS", "VALLEYS",
		},
	},
	{
		Primary: "VIADUCT",
		Short:   "VIA",
		Alt: []string{
			"VIA", "VDCT", "VIADCT", "VIADUCT",
		},
	},
	{
		Primary: "VIEW",
		Short:   "VW",
		Alt: []string{
			"VW", "VIEW",
		},
	},
	{
		Primary: "VIEWS",
		Short:   "VWS",
		Alt: []string{
			"VWS", "VIEWS",
		},
	},
	{
		Primary: "VILLAGE",
		Short:   "VLG",
		Alt: []string{
			"VLG", "VILL", "VILLG", "VILLAG", "VILLAGE", "VILLIAGE",
		},
	},
	{
		Primary: "VILLAGES",
		Short:   "VLGS",
		Alt: []string{
			"VLGS", "VILLAGES",
		},
	},
	{
		Primary: "VILLE",
		Short:   "VL",
		Alt: []string{
			"VL", "VILLE",
		},
	},
	{
		Primary: "VISTA",
		Short:   "VIS",
		Alt: []string{
			"VIS", "VIST", "VISTA", "VST", "VSTA",
		},
	},
	{
		Primary: "WALK",
		Short:   "WALK",
		Alt: []string{
			"WALK", "WALKS",
		},
	},
	// NOTE: merged WALKS with WALK because the unstable
	// normalization can't be supported
	// {
	// 	Primary: "WALKS",
	// 	Short: "WALK",
	// 	Alt: []string{
	// 		"WALKS",
	// 	},
	// },
	{
		Primary: "WALL",
		Short:   "WALL",
		Alt: []string{
			"WALL",
		},
	},
	{
		// USPS preferred abbreviation is WAY (not WY). WY remains an accepted
		// input alt so "… Way" / "… WY" both normalize to StreetSuffix WAY.
		// This also disambiguates state-as-name "WY WY" → WYOMING WAY.
		Primary: "WAY",
		Short:   "WAY",
		Alt: []string{
			"WY", "WAY",
		},
	},
	{
		Primary: "WAYS",
		Short:   "WAYS",
		Alt: []string{
			"WAYS",
		},
	},
	{
		Primary: "WELL",
		Short:   "WL",
		Alt: []string{
			"WL", "WELL",
		},
	},
	{
		Primary: "WELLS",
		Short:   "WLS",
		Alt: []string{
			"WLS", "WELLS",
		},
	},
}

var streetSuffixPrimaryMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range streetSuffixes {
		for _, a := range v.Alt {
			if !yield(a, v.Primary) {
				return
			}
		}
	}
})
var streetSuffixShortMap = maps.Collect(func(yield func(string, string) bool) {
	for _, v := range streetSuffixes {
		for _, a := range v.Alt {
			if !yield(a, v.Short) {
				return
			}
		}
	}
})
var streetSuffixKeys = slices.Collect(maps.Keys(streetSuffixPrimaryMap))

var alphaspace = regexp.MustCompile("[^a-zA-Z ]+")

func normalizeStreetSuffix(r string, primary bool, fuzzy bool) (string, error) {
	// clean out any punctuation
	clean := alphaspace.ReplaceAllString(r, "")
	// capitalize
	capitalized := strings.ToUpper(clean)
	// if requested, fuzzy match keys
	rkey := capitalized
	if fuzzy && len(capitalized) > 3 {
		matched, err := edlib.FuzzySearchThreshold(capitalized, streetSuffixKeys, 0.7, edlib.DamerauLevenshtein)
		if err != nil {
			// TODO: figure out how to let the user control logging in this library
			// log warn "Unable to fuzzy match supplied region string"
			matched = capitalized
		} else {
			rkey = matched
		}
	}

	if primary {
		// look up the primary
		primary, ok := streetSuffixPrimaryMap[rkey]
		if !ok {
			return "", fmt.Errorf("Unrecognized street suffix: %s", rkey)
		}

		return primary, nil
	}

	// look up the abbreviation
	abrev, ok := streetSuffixShortMap[rkey]
	if !ok {
		return "", fmt.Errorf("Unrecognized street suffix: %s", rkey)
	}

	return abrev, nil
}

func NormalizeStreetSuffix(r string) (string, error) {
	return normalizeStreetSuffix(r, true, false)
}

func FuzzyNormalizeStreetSuffix(r string) (string, error) {
	return normalizeStreetSuffix(r, true, true)
}

func NormalizeStreetSuffixAbreviation(r string) (string, error) {
	return normalizeStreetSuffix(r, false, false)
}

func FuzzyNormalizeStreetSuffixAbreviation(r string) (string, error) {
	return normalizeStreetSuffix(r, false, true)
}
