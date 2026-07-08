package goprojectusat

// split a candidate address at lines and commas
// - account for location of splits
// split a candidate address at spaces
// filter out empty segments
// flag segments for possible address parts
// - country is always last, if it exists
// - postal code is:
//   - last before country
//   - 5 digits or 5 digits plus 4 digits after a hyphen when country is USA (including unspecified, mostly)
//   - Canada and other international postal codes may throw this off
//     - but still likely alphanumeric with hyphens
// - region (state, possession, canadian province, military command)
//   - last before postal code
//   - should be in a defined set of names for US and Canada
//     - might use lowest levenshtein distance to support misspellings
//       (see https://github.com/hbollon/go-edlib#2-most-matching-unique-result-with-threshold)
// - city
//   - directly before region
//   - no numbers (any exceptions?)
//   - often starts a new line (not always!!!)
// - secondary unit
//   - may not exist
//   - likely contains or follows an indicator word or symbol
//   - often alphanumeric with possible hyphen or forward slash
//   - very likely to be after the street (after a street suffix)
//   - may be on a line after the street
//
// - postdirectional
//
// - street
//   - very likely to contain a street suffix
//   - usually not numeric, but can be (7th St.)
//   - often multiple words
//
// - predirectional
//
// - primary address number
//   - may be alphanumeric, usually numeric
//   - may contain "." or "/"
//   - does not look like "1st" or other potential street names
//
// - addressee, buisiness name, etc
//   - first line(s)
//     - is there a later segment that looks like a street address?
//   - might have signals like "Care of", "Attention", "Inc."
//   - doesn't look like it starts with a street number
//   - mostly text without a street suffix (usually)
