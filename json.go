//  Copyright (c) 2015 Couchbase, Inc.

package gson

// SpaceKind to skip white-spaces in JSON text.
type SpaceKind byte

const (
	// AnsiSpace will skip white space characters defined by ANSI spec.
	AnsiSpace SpaceKind = iota + 1

	// UnicodeSpace will skip white space characters defined by Unicode spec.
	UnicodeSpace
)

func (ws SpaceKind) String() string {
	switch ws {
	case AnsiSpace:
		return "AnsiSpace"
	case UnicodeSpace:
		return "UnicodeSpace"
	default:
		panic("new space-kind")
	}
}
