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

Since this is a single value type, and all null values shall be encoded
into as:

```go
$ ./gson -inptxt null -json2collate
Json: null
Coll: "2\x00"
Coll: [50 0]
```

Null values will sort before any other JSON value.

Boolean
-------

There are two values, `false` and `true`. Value `true` will sort after
value `false`, while value `false` will always sort after JSON null value.
Boolean false and true shall be encoded as:

```shell
./gson -inptxt false -json2collate
Json: false
Coll: "<\x00"
Coll: [60 0]
$ ./gson -inptxt true -json2collate
Json: true
Coll: "F\x00"
Coll: [70 0]
```

Number
------

Number is the trickiest element of all JSON types. To begin with, JSON
specification does not define an lower or upper bound for numbers.
And many implementation treat JSON numbers as float64 type. But
fortunately we have strong conception of number sequence. There are two
choices that can be taken when encoding number:

* Numbers can be treated as **float64** defined by IEEE-754 specification.
  This implies that, only integers less than 2^53 and greater than -2^53
  can be represented using this format.
* JSON parsers can automatically detect integers outside 2^53 boundary.

The trade-off between the two is that, former implementation may consume less
CPU while later implementation will be future proof. Either way, we will
be encoding all number as arbitrarily sized floating point.

A floating point number `f` takes a mantissa `m` and an integer exponent
`e` such that `f = 10e × ±m`. Unlike the IEEE-754 specification mantissa and
exponent can be of arbitrary length which helps us to encode very larger
numbers. The mantissa will always have a non-zero digit after
the point and for each number `f` will have a unique exponent. This is the
normalised representation and ensures comparison can be determined by the
exponent where it differs otherwise by the mantissa.

The floating point number is then a triple `(+, e, m)` or `(−, −e, m)`
consisting of a sign followed by the exponent and mantissa. The exponent
is represented as an integer using the recursive method and is negated for
negative numbers to ensure ordering. The mantissa is represented as a
positive small decimal but its sign must appear at the front of the entire
representation. As with large decimals there is no need to include a
delimiter since the exponent is length prefixed. Zero is represented as
0 since a sign symbol will be used for all other numbers. The resulting
representations look like the large decimals with a sign added except the
integer is used multiplicatively to scale the number rather than as an
additive offset. Following is a set of examples:

```text
float           binary compiled

−0.1 × 10^11    - --7888+
−0.1 × 10^10    - --7898+
-1.4            - -885+
-1.3            - -886+
-1              - -88+
-0.123          - 0876+
-0.0123         - +1876+
-0.001233       - +28766+
-0.00123        - +2876+
0               0
+0.00123        + -7123-
+0.001233       + -71233-
+0.0123         + -8123-
+0.123          + 0123-
+1              + +11-
+1.3            + +113-
+1.4            + +114-
+0.1 × 10^10    + ++2101-
+0.1 × 10^11    + ++2111-
```

String
------

ASCII formatted strings are similar to binary comparison of byte arrays.
Every character in ASCII has a corresponding byte value and byte order has
one-one correspondence with character ordering. This get complicated once
we move on to Unicoded strings.

Strings are collated as it is received from the input **without un-quoting**
as JSON-string and **without unicode collation**. Encoded strings shall be
byte stuffed to escape item Terminator.

```shell
$ ./gson -inptxt '"hello\x00world"' -json2collate
Json: "hello\x00world"
Coll: "Zhello\x00\x0100world\x00\x00"
Coll: [90 104 101 108 108 111 0 1 48 48 119 111 114 108 100 0 0]
```

Array
-----

Sorting arrays have two aspects to it.

One, we should compare each item from both the array one after the other in
their positional order. If items are of different types then we should
follow the sort order between types.

Second aspect of sorting array is its arity, whether array with larger number
of items should sort after array with smaller number of items. If arity of
array needs to be considered, then we shall compare the length of each array
before comparing each item from both the array.

```shell
$ ./gson -inptxt "[10,true,null]" -json2collate
Json: [10,true,null]
Coll: "nP>>21-\x00F\x002\x00\x00"
Coll: [110 80 62 62 50 49 45 0 70 0 50 0 0]
```

**For partial compare between two binary-compiled arrays, prune the last byte
from the encoded text of smaller array and continue with binary comparison.**

Property
--------

Sorting arrays have three aspects to it.

A property object is made up `{key,value}` pairs, where key must be a JSON
string, and value can be of any JSON type. For the purpose of sorting, we
shall first sort the {key,value} pairs within a property object based on
the `key` string.

Secondly, we pick each {key,value} pair from both the property in its
positional order and start the comparison, `key` is compared first
and `value` is compared only when `key` from each item compares equal.

Third aspect of sorting property is its arity, whether property with
larger number of items should sort after property with smaller number
of items. If arity of property needs to be considered, then we shall
compare the length of each property before comparing each `{key,value}`
item from both the property.

```shell
$ ./gson -inptxt '{"first":true, "second":false}' -json2collate
Json: {"first":true, "second":false}
Coll: "xd>2\x00Zfirst\x00\x00F\x00Zsecond\x00\x00<\x00\x00"
Coll: [120 100 62 50 0 90 102 105 114 115 116 0 0 70 0 90 115 101 99 111 110 100 0 0 60 0 0]
```
