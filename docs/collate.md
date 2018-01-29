Sorting JSON data, crazy fast
=============================

Collation is sorting data in relation to each other. That is, to figure
out what comes before and what comes after.

Let us begin with numbers, more specifically let us try to sort a set of
numbers: `{100, 346, 20, 560}`. It is quite obvious how the sorted set
of numbers will look like: `{20, 100, 346, 560}`. For instance
following snippet of program, also called as bubble sort, can do the
sorting business for us:

```go
func bubblesort() {
    list := []int{100, 346, 20, 560}
    for itemCount := len(list) - 1; ; itemCount-- {
        hasChanged := false
        for index := 0; index < itemCount; index++ {
            if a[index] > a[index+1] {
                a[index], a[index+1] = a[index+1], a[index]
                hasChanged = true
            }
        }
        if hasChanged == false {
            break
        }
    }
}
```

In the above program, `a[index] > a[index+1]` is the comparator to sort
our number set. This expression will be compiled into machine instruction
that can compare two 64-bit integers, provided the operands to comparator
are of type integer and hence stored either in little-endian or big-endian
binary format (depending on the processor). Will above comparator expression
or its corresponding machine instruction work if our data type is supplied
like this,

```go
list := []string{"100", "346", "20", "560"}
```

So what is the difference ? In the second case we are supplying the integers
in ASCII format, or TEXT format, or more specifically as string type. Can
we use the same comparator expression, that we used for integer type ?
Some languages, especially those that are dynamically typed can automagically
pick the right comparator function, at runtime, based on the type of its
operands. Nevertheless, the comparator functions are going to be distinctly
different and so its output.

```python
> list = ["100", "346", "20", "560"]
> list.sort()
> list
['100', '20', '346', '560']
```

The comparator function used to sort a list of numbers represented as string
type is not really working out. What went wrong ? Our mistake was in
choosing a wrong comparator function by choosing a wrong data representation.

**This is the central problem of Collating JSON data. All data in JSON format
are meant to be in human readable text format**

Internet, Web and JSON
----------------------

Web is the human view of internet and JSON is the text representation of
of web data.

Every data type, be it simple types like number, string, boolean, and
nil or composite types like array and property-object are all represented
in text for human consumption which, unfortunately, machines cannot
understand without parsing that text representation into binary
representation. To get an idea on how costly it is, let us try an experiment
to sort numbers in JSON format and compare it with native integers.

First let us try sorting a large set of native integers:

```python
ls = list(reversed(range(1000000)))
start = time.time()
ls.sort()
print("took %s" % (time.time() - start))
// output on a 2015 model mac-book-pro
// took 0.0256040096283 seconds
```

Next let us try sorting a large set of integers in text (JSON) representation,
our program uses python's magic of dynamic programming to override comparator
function used by `list.sort`.

```python
import time

class I(object):
    def __init__(self, value) :
        self.value = value

    def __lt__(self, other):
        return int(self.value) < int(other.value)

    def __repr__(self) :
        return self.value

ls = [I(str(x)) for x in reversed(range(1000000)) ]
start = time.time()
ls.sort()
print("took %s seconds" % (time.time() - start))
// output on a 2015 model mac-book-pro
// took 1.69596195221 seconds
```

**Sorting 1 Million numbers in text representation takes 66 times more
CPU than sorting 1 Million machine native 64-bit integers**.

Sorting data in JSON format
===========================

Our objective is to compile data from JSON format to a binary format
without losing any information contained in the data. We have two goals
to achive in doing so:

* Preserve the sort order of input text. That is, after we compile a
  set of JSON encoded data into a binary format, we should be able use
  **memcmp** as the comparator function to sort encoded data, and the
  sorted set should preserve the sort order as if they are compared
  by parsing the JSON text.
* Binary compiled output shall be decoded back to JSON text without losing
  any information.

To sort data we need a strong idea of what come before and what comes
after. Before working out a sort order for each JSON type let us
work out a sort order between all JSON types:

```go
Terminator  byte = 0

TypeNull    byte = 50
TypeFalse   byte = 60
TypeTrue    byte = 70
TypeNumber  byte = 80
TypeString  byte = 90
TypeLength  byte = 100
TypeArray   byte = 110
TypeObj     byte = 120
TypeBinary  byte = 130
```

After compiling the JSON value into binary format it is prefixed with
one-byte type value, and suffixed by Terminator byte. A list of such
type prefix is provided above. Although `TypeTrue` and `TypeFalse` are
of same type, to save memory footprint, we shall include them in type
ordering as to distinct types.

Sort order for each JSON type:

Null
----

This type shall be encoded with the type header.

**bool**

There are two values, `false` and `true`. We shall assume that value
`false` will sort before value `true`.

**number**

Number is the trickiest element of all JSON types. For one, JSON
specification does not define an lower or upper bound for numbers.
And many implementation treat JSON numbers as float64 type. But
fortunately we have strong conception of number sequence.

**string**

ASCII formated strings are similar to binary comparision of byte
arrays. Every character in ASCII has a corresponding byte value and
byte order has one-one correspondence with character ordering. This
get complicated once we move on to unicoded strings.

**Array**

Sorting arrays have two aspects to it. One, we should compare each item
from both the array one after the other in their positional order. If
items are of different types then we should follow the sort order between
types. Second aspect of sorting array is its arity, whether array with
larger number of items should sort after array with smaller number of
items. If arity of array needs to be considered, then we shall compare
the length of each array before comparing each item from both the array.

**Property**

Sorting arrays have three aspects to it. A property object is made up
`{key,value}` pairs, where key must be a JSON string, and value can be
of any JSON type. For the purpose of sorting, we shall first sort the
{key,value} pairs within a property object based on the `key` string.
Secondly, we pick each {key,value} pair from both the property in its
positional order and start the comparision, `key` is compared first
and `value` is compared only when `key` from each item compares equal.
Thirdly aspect of sorting property is its arity, whether property with
larger number of items should sort after property with smaller number
of items. If arity of property needs to be considered, then we shall
compare the length of each property before comparing each `{key,value}`
item from both the property.

Encoding Nil, Boolean
---------------------

Types `nil, true, and false` are encoded as single byte. Types `string`
value will be encoded into UTF-8 format before collating it.

**Encoding Number**

The basic problem with numbers is that Javascript, from which JSON evolved,
does not have any notion of integer numbers. All numbers are represented as
64-bit floating point. In which case, only integers less than 2^53 and greater
than -2^53 can be represented using this format.

**Encoding golang values**

* Types `nil`, `true`, `false`, `float64`, `int64`, `int`,
  `string`, `[]byte`, `[]interface{}`, `map[string]interface{}`
  are supported for collation.
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
