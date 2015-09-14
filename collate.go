package gson

// while encoding JSON data-element, both basic and composite, encoded
// string is prefixed with a type-byte. `Terminator` terminates encoded
// datum.
const (
	Terminator byte = iota
	TypeMissing
	TypeNull
	TypeFalse
	TypeTrue
	TypeNumber
	TypeString
	TypeLength
	TypeArray
	TypeObj
	TypeBinary
)

// Length is an internal type used for prefixing collated arrays
// and properties with number of items.
type Length int64

// Missing denotes a special type for an item that evaluates
// to _nothing_, used for collation.
type Missing string

// MissingLiteral is special string to denote missing item.
// IMPORTANT: we are assuming that MissingLiteral will not
// occur in the keyspace.
const MissingLiteral = Missing("~[]{}falsenilNA~")

// Equal checks wether n is MissingLiteral
func (m Missing) Equal(n string) bool {
	s := string(m)
	if len(n) == len(s) && n[0] == '~' && n[1] == '[' {
		return s == n
	}
	return false
}
