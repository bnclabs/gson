--------
Summary:
--------

Primary use of collatejson is to do binary comparison on two json strings.
Binary comparison (aka memcmp) can be several times faster than any other
custom JSON parser.

-----
JSON:
-----

JSON, as defined by spec, can represent following elements,
  1. nil.
  2. boolean.
  3. numbers in integer form, or floating-point form.
  4. string.
  5. array of JSON elements.
  6. object of key,value properties, where key is represetned as string and
     value can be any JSON elements.

Nil, Boolean:
-------------

nil, true, and false are encoded as single byte.

Number:
-------

The basic problem with numbers is that Javascript, from which JSON evolved,
does not have any notion of integer numbers. All numbers are represented as
64-bit floating point. In which case, only integers less than 2^53 and greater
than -2^53 can be represented using this format.

* collating golang values:

  * all number are collated as float.
  * all JSON float64 numbers are collated as float64, and,
    64-bit integers > 2^53 are preseved as integer and collated as float.
  * array-length (if configured) and property-length (if configured) are
    collated as integer.

* collating JSON encoded numbers:

  * all number are collated as float.
  * if config.nk is FloatNumber, all numbers are interpreted as float64
    and collated as float64.
  * if config.nk is SmartNumber, all JSON float64 numbers are collated as
    float64, and, 64-bit integers > 2^53 are preseved as integer and collated
    as float.
  * array-length (if configured) and property-length (if configured) are
    collated as integer.

* collating CBOR encoded numbers:

  * all number are collated as float.
  * if config.nk is FloatNumber, 64-bit integers are converted to float64.
  * if config.nk is SmartNumber, 64-bit integers > 2^53 and < -2^53 are
    preseved as integer and collated as float, without loosing it loosing
    precision.
  * array-length (if configured) and property-length (if configured) are
    collated as integer.

* transforming collated numbers to CBOR encoded numbers:

  * since all numbers are collated as float, it is converted back to text
    representation of float, in format: [+-]x.<mantissa>e[+-]<exp>.
  * if config.nk is FloatNumber, all number are encoded as CBOR-float64.
  * if config.nk is SmartNumber, all numers whose exponent is >= 15 is encoded
    as uint64 (if number is positive), or int64 (if number is negative).
    Others are encoded as CBOR-float64.

* transforming collated numbers to JSON encoded numbers:

  * since all numbers are collated as float, it is converted back to text
    representation of float, in format: [+-]x.<mantissa>e[+-]<exp>.
  * if config.nk is FloatNumber, all number are encoded as JSON-float64.
  * if config.nk is SmartNumber, all numers whose exponent is >= 15 is encoded
    as uint64 (if number is positive), or int64 (if number is negative).
    Others are encoded as JSON-float64.

* transforming collated numbers to value:

  * since all numbers are collated as float, it is converted back to text
    representation of float, in format: [+-]x.<mantissa>e[+-]<exp>.
  * if config.nk is FloatNumber, all number are encoded as JSON-float64.
  * if config.nk is SmartNumber, all numers whose exponent is >= 15 is encoded
    as uint64 (if number is positive), or int64 (if number is negative).
    Others are treated as float64.

* transforming JSON encoded numbers to CBOR-numbers:

  * if config.nk is FloatNumber, all numbers are encoded as CBOR-float64.
  * if config.nk is SmartNumber, all JSON float64 numbers are encoded as
    CBOR-float64, and, all JSON positive integers are encoded as
    CBOR-uint64, and, all JSON negative integers are encoded as
    CBOR-int64.

* transforming JSON encoded numbers to golang values:

  * if config.nk is FloatNumber, all numbers are interpreted as float64.
  * if config.nk is SmartNumber, all JSON integers are interpreted as either
    uint64 or int64, and, JSON float64 are interpreted as float64.

Array:
------

By default array is not prefixed with length of the array, which means
elements are compared one by one until binary-compare returns EQ, GT or
LT. This is assuming that elements in both arrays (key1 and key2) have
one-to-one correspondence with each other. **Suppose number of elements
in key1 is less that key2, or vice-versa, prune the last byte from the
encoded text of smaller array and continue with binary comparison.**

Object:
-------

By default objects are prefixed with length of the object (ie) number of
elements in the object. This means objects with more number of {key,value}
properties will sort after.

While encoding collatejson will sort the object properties based on keys.
This means the property key will be compared first and if equal, comparison
will continue to its value.

Note:

1. Wildcards are not accepted in elements. For instance, it is not possible to
   select all Cities starting with "San". To get all cities starting with
   "San" perform a ">=" operation on storage and stop iterating when returned
   value is not prefixed with "San"
