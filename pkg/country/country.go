package country

// country is always the last line of an address in a multiline address according to US Project@ rules.
// Commas should be interpreted similarly. In a signle line address without commas, we may parse the
// country as the set of tokens containing only letters after what looks like a zip code.
