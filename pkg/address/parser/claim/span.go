package claim

// Span is a run of tokens, in the same index space claims are made in.
//
// It lives here rather than in the package that first needed it because it is
// the coordinate system, not a feature of any one consumer. A last line extent
// and a run of tokens with nowhere to live are the same kind of thing said by
// different packages, and they have to be comparable.
//
// A Span is not a Claim and deliberately carries no meaning. What it is for is
// naming tokens a reading does not explain — a claim says what tokens are, and
// there has to be a way to say that some tokens are nothing yet.
type Span struct {
	Start  int
	Length int
}

// End returns the index one past the last token in the span.
func (s Span) End() int {
	return s.Start + s.Length
}

// Gaps returns the runs of tokens within span that no part covers.
//
// The parts need not be sorted and may come from more than one claim, which is
// the normal case: what a reading leaves unexplained is a question about the
// reading as a whole, not about any one claim in it. Parts lying outside the
// span are ignored rather than treated as an error, so a caller can pass
// everything it accepted and ask about a window of it.
func (s Span) Gaps(parts []ClaimPart) []Span {
	covered := make([]bool, max(0, s.Length))

	for _, p := range parts {
		for i := max(p.Start, s.Start); i < min(p.End(), s.End()); i++ {
			covered[i-s.Start] = true
		}
	}

	var gaps []Span
	start := -1

	for i := range covered {
		switch {
		case !covered[i] && start < 0:
			start = i
		case covered[i] && start >= 0:
			gaps = append(gaps, Span{s.Start + start, i - start})
			start = -1
		}
	}

	if start >= 0 {
		gaps = append(gaps, Span{s.Start + start, len(covered) - start})
	}

	return gaps
}
