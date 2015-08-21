package collate

// while encoding JSON data-element, both basic and composite, encoded string
// is prefixed with a type-byte. `Terminator` terminates encoded datum.
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
)
