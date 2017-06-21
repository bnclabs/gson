Summary
=======

Collation is sorting data in relation to each other. That is, to figure
out what comes before and what comes after. Machine (Language) representation
of data can be a number, string, null, array, map etc.. To sort data that
come in one or more of the machine representation, we need a strong conception
of what comes before and what comes after.

Sort order
==========

Following is an example of sort order between different data types.

```go
	TypeNull    byte = 2
	TypeFalse   byte = 3
	TypeTrue    byte = 4
	TypeNumber  byte = 5
	TypeString  byte = 6
	TypeLength  byte = 7
	TypeArray   byte = 8
	TypeObj     byte = 9
	TypeBinary  byte = 10
```

Here type ``null`` sorts before than every other type.

Binary compilation
==================

If input text, which is in JSON encoded format, needs to be sorted based
on above collation rules, every comparison operation will involve parsing
the input for its type and value. And hence comparison is a costly operation.

What if we can compile the input text into binary sequence of bytes with
following properties ?

- Preserve the sort order of input text. That is, when a set of binary
compiled input text, sorted using **memcmp**, preserves the sort order
as if they are compared by parsing the JSON text.
- Binary compile output can be decoded back to JSON text without loss
of any information.

With [gson](http://github.com/prataprc/gson) we are trying to achieve
precisely that.

Encoding Nil, Boolean
---------------------

nil, true, and false are encoded as single byte.

**Encoding Number**

The basic problem with numbers is that Javascript, from which JSON evolved,
does not have any notion of integer numbers. All numbers are represented as
64-bit floating point. In which case, only integers less than 2^53 and greater
than -2^53 can be represented using this format.

**Encoding golang values**

* All number are collated as float.
* All JSON float64 numbers are collated as float64, and,
  64-bit integers > 2^53 are preserved as integer and collated as float.
* Array-length (if configured) and property-length (if configured) are
  collated as integer.

**Encoding JSON numbers**

* All number are collated as float.
* If config.nk is FloatNumber, all numbers are interpreted as float64
  and collated as float64.
* If config.nk is SmartNumber, all JSON float64 numbers are collated as
  float64, and, 64-bit integers > 2^53 are preserved as integer and collated
  as float.
* Array-length (if configured) and property-length (if configured) are
  collated as integer.

**Encoding CBOR numbers**

* All number are collated as float.
* If config.nk is FloatNumber, 64-bit integers are converted to float64.
* If config.nk is SmartNumber, 64-bit integers > 2^53 and < -2^53 are
  preserved as integer and collated as float, without loosing its
  precision.
* Array-length (if configured) and property-length (if configured) are
  collated as integer.

**Decoding to CBOR numbers**

* Since all numbers are collated as float, it is converted back to text
  representation of float, in format: [+-]x.<mantissa>e[+-]<exp>.
* If config.nk is FloatNumber, all number are encoded as CBOR-float64.
* If config.nk is SmartNumber, all numbers whose exponent is >= 15 is encoded
  as uint64 (if number is positive), or int64 (if number is negative).
  Others are encoded as CBOR-float64.

**Decoding to JSON numbers**

* Since all numbers are collated as float, it is converted back to text
  representation of float, in format: [+-]x.<mantissa>e[+-]<exp>.
* If config.nk is FloatNumber, all number are encoded as JSON-float64.
* If config.nk is SmartNumber, all numers whose exponent is >= 15 is encoded
  as uint64 (if number is positive), or int64 (if number is negative).
  Others are encoded as JSON-float64.

**Decoding to golang value**

* Since all numbers are collated as float, it is converted back to text
  representation of float, in format: [+-]x.<mantissa>e[+-]<exp>.
* If config.nk is FloatNumber, all number are encoded as JSON-float64.
* If config.nk is SmartNumber, all numers whose exponent is >= 15 is encoded
  as uint64 (if number is positive), or int64 (if number is negative).
  Others are treated as float64.

String
------

Strings are collated as it is received from the input **without
un-quoting** as JSON-string and **without unicode collation**.
Encoded strings are byte stuffed to escape item Terminator.

Array:
------

By default array is not prefixed with length of the array, which means
elements are compared one by one until binary-compare returns EQ, GT or
LT. This is assuming that elements in both arrays (key1 and key2) have
one-to-one correspondence with each other.

If config.SortbyArrayLen() is true, arrays having more number of items
sort after arrays having lesser number of items. Two arrays of same
arity will follow the same procedure as described above.

**Suppose number of elements in key1 is less that key2, or vice-versa,
then prune the last byte from the encoded text of smaller array and
continue with binary comparison.**

Object:
-------

By default objects are prefixed with length of the object (ie) number of
elements in the object. This means objects with more number of {key,value}
properties will sort after.

While encoding collatejson will sort the object properties based on keys.
This means the property key will be compared first and if equal, comparison
will continue to its value.

If config.SortbyPropertyLen() is false, then keys of each object are sorted
and key from each maps is matched in the sort order, after which values
will be matched. Note that sorting of keys and encoding of keys and values
shall be done during encoding time.
