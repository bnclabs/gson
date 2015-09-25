//  Copyright (c) 2015 Couchbase, Inc.

package gson

// Collation order for supported types, to change the order set these
// values in your init() function.
var (
	Terminator  byte = 0
	TypeMissing byte = 1
	TypeNull    byte = 2
	TypeFalse   byte = 3
	TypeTrue    byte = 4
	TypeNumber  byte = 5
	TypeString  byte = 6
	TypeLength  byte = 7
	TypeArray   byte = 8
	TypeObj     byte = 9
	TypeBinary  byte = 10
)

// Length is an internal type used for prefixing collated arrays
// and properties with number of items.
type Length int64

// Missing denotes a special type for an item that evaluates
// to _nothing_, used for collation.
type Missing string

// MissingLiteral is special string to denote missing item:
//	IMPORTANT: we are assuming that MissingLiteral will not occur in
//	the keyspace.
const MissingLiteral = Missing("~[]{}falsenilNA~")

// Equal checks wether n is MissingLiteral
func (m Missing) Equal(n string) bool {
	s := string(m)
	if len(n) == len(s) && n[0] == '~' && n[1] == '[' {
		return s == n
	}
	return false
}
