[![Build Status](https://travis-ci.org/prataprc/gson.png)](https://travis-ci.org/prataprc/gson)
[![Coverage Status](https://coveralls.io/repos/prataprc/gson/badge.png?branch=master&service=github)](https://coveralls.io/github/prataprc/gson?branch=master)
[![GoDoc](https://godoc.org/github.com/prataprc/gson?status.png)](https://godoc.org/github.com/prataprc/gson)


What is what
------------

**json**

* Java Script Object Notation, also called [JSON](http://www.json.org/),
  [RFC-7159](https://tools.ietf.org/html/rfc7159).
* fast becoming the internet standard for data exchange.
* human readable format, not so friendly for machine representation.

**value (aka gson)**

* golang object parsed from json, cbor or collate representation.
* json arrays are represeted in golang as `[]interface{}`.
* json objects, aka properties, are presented in golang as
  `map[string]interface{}`.
* gson objects support operations like, Get(), Set(), and
  Delete() on individual fields located by the json-pointer.

**cbor**

* Concise Binary Object Representation, also called [CBOR](http://cbor.io/),
  [RFC-7049](https://tools.ietf.org/html/rfc7049).
* machine friendly, designed for inter-networking of light weight
  devices, and easy to implement in many languages.
* can be used for more than data exchange, left to user
  imagination :) ...

**collate**

* a custom encoding based on a paper and improvised to handle
  JSON specification.
* binary representation preserving the sort order.
* transform back to original JSON from binary representation.
* numbers can be encoded in three ways - as integers, or as
  small-decimals, or as floating-point represented in exponent
  form.
* strings are collated as it is received from the input **without
  un-quoting** the JSON-string and **without unicode collation**.
* strings are byte stuffed to escape item Terminator.
* if config.SortbyArrayLen() is true, arrays having more number
  of items sort after arrays having lesser number of items.
* if config.SortbyPropertyLen() is true, properties having more
  number of items sort after properties having lesser number of
  items.
* items in a property object are sorted by its property name
  before they are compared with property's value.
* per couchbase-N1QL requirement: collation of missing type, a
  unit type, is also supported.

**json-pointer**

* url like field locator within a json object,
  [RFC-6901](https://tools.ietf.org/html/rfc6901).
* make sense only for json arrays and objects, but to any level
  of nesting.
* json-pointers shall be unquoted before they are used for
  accessing into json text (or an equivalent representation),
  after unquoting segments within the pointer, each segment shall
  be binary compared with property keys.
* json-pointers can be used to access gson or cbor representation.
* documents encoded in cbor format using LengthPrefix are not
  supported by lookup APIs.

Transforms
----------

![transforms](docs/transforms.png)

**json to value**

* a custom parser is supplied that must be faster than encoding/json.
* numbers can be interpreted as integer, or float64, or retained as
  string based on the configuration parameter `NumberKind`.
  *  `JsonNumber` to retain number as JSON string type aliased
     to `encoding/json.Number`.
  *  `IntNumber` to interpret JSON number as integer whose size is
     defined by the platform.
  *  `FloatNumber` to interpret JSON number as 64-bit floating point.
* whitespace can be interpreted, based on configuration parameter
  `SpaceKind`, as `AnsiSpace` that should be faster
  than `UnicodeSpace`, while the later supports unicode whitespaces
  as well.

```text
BenchmarkJson2ValFlt    10000000         161 ns/op    55.66 MB/s           8 B/op        1 allocs/op
BenchmarkJson2ValJsn    10000000         121 ns/op    74.06 MB/s          16 B/op        1 allocs/op
BenchmarkJson2ValString  3000000         467 ns/op    68.52 MB/s          80 B/op        3 allocs/op
BenchmarkJson2ValArr5    1000000        1013 ns/op    29.60 MB/s         719 B/op        8 allocs/op
BenchmarkJson2ValMap5     500000        2998 ns/op    20.67 MB/s        5166 B/op       14 allocs/op
BenchmarkJson2ValTyp       50000       24049 ns/op    60.50 MB/s       17613 B/op      127 allocs/op
```

as compared to using encoding/json for the same data sample:

```text
BenchmarkUnmarshalFlt    1000000        1389 ns/op    6.48 MB/s        264 B/op        3 allocs/op
BenchmarkUnmarshalNum    1000000        1401 ns/op    6.42 MB/s        264 B/op        3 allocs/op
BenchmarkUnmarshalStr    1000000        1939 ns/op   16.50 MB/s        336 B/op        4 allocs/op
BenchmarkUnmarshalArr5    500000        3329 ns/op    9.01 MB/s        320 B/op       10 allocs/op
BenchmarkUnmarshalMap5    200000        8715 ns/op    7.11 MB/s        976 B/op       41 allocs/op
BenchmarkUnmarshalTyp      20000       66649 ns/op   21.83 MB/s       6544 B/op      248 allocs/op
```

**value to json**

* to convert value back to json text golang's encoding/json package is
  used.
* `Encoder` interface{} is used to re-use o/p buffer.

```
BenchmarkVal2JsonFlt     3000000        561 ns/op   16.02 MB/s         8 B/op        1 allocs/op
BenchmarkVal2JsonString  2000000        631 ns/op   60.20 MB/s         8 B/op        1 allocs/op
BenchmarkVal2JsonArr5    1000000       1470 ns/op   19.72 MB/s        72 B/op        6 allocs/op
BenchmarkVal2JsonMap5     300000       5025 ns/op   10.15 MB/s       601 B/op       23 allocs/op
BenchmarkVal2JsonTyp       50000      30180 ns/op   39.33 MB/s      3488 B/op      141 allocs/op
```

**value to cbor**

* `nil`, `true`, `false` golang types are encodable into cbor
  format.
* all golang `number` types, including signed, unsigned, and
  floating-points variants, are encodable into cbor format.
* `[]byte` is encoded as cbor byte-string.
* `string` is encoded as cbor text.
* generic `array` is interpreted as golang `[]interface{}` and
  encoded as cbor array.
  * with `LengthPrefix` option for ContainerEncoding, arrays and
    maps are encoded with its length.
  * with `Stream` option, arrays and maps are encoded using
    Indefinite and Breakstop encoding.
* generic `property` is interpreted as golang `[][2]interface{}`
  and encoded as cbor array of 2-element array, where the first item
  is key represented as string and second item is any valid json
  value.
* before encoding `map[string]interface{}` type, use
  `GolangMap2cborMap()` function to transform them to
  `[][2]interface{}`.
* following golang data types are encoded using cbor-tags,
  * `time.Time` encoded with tag-0.
  * `Epoch` type supplied by cbor package, encoded
    with tag-1.
  * `EpochMicro` type supplied by cbor package, encoded
    with tag-1.
  * `math/big.Int` positive numbers are encoded with tag-2, and
    negative numbers are encoded with tag-3.
  * `DecimalFraction` type supplied by cbor package,
    encoded with tag-4.
  * `BigFloat` type supplied by cbor package, encoded
    with tag-5.
  * `Cbor` type supplied by cbor package, encoded with
    tag-24.
  * `regexp.Regexp` encoded with tag-35.
  * `CborPrefix` type supplied by cbor package, encoded
    with tag-55799.
* all other types shall cause a panic.

```text
BenchmarkVal2CborNull   200000000    8.75 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborTrue   100000000   18.0 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborFalse  100000000   18.6 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborUint64 50000000    24.2 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborInt64  50000000    26.7 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborFlt32  100000000   20.8 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborFlt64  50000000    25.7 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborBytes  50000000    34.7 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborText   30000000    42.5 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CborArr5   10000000   175 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborMap5    3000000   414 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborTyp      500000  3941 ns/op     0 B/op  0 allocs/op
```


**cbor to value**

* reverse of all `value to cbor` encoding, described above, are
  supported.
* cannot decode `float16` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks shall be decoded outside
  this package using `IsIndefinite*()` and `IsBreakstop()` APIs.

```text
BenchmarkCbor2ValNull	50000000     37.1 ns/op	   0 B/op	 0 allocs/op
BenchmarkCbor2ValTrue	20000000     67.9 ns/op	   1 B/op	 1 allocs/op
BenchmarkCbor2ValFalse	20000000     68.9 ns/op	   1 B/op	 1 allocs/op
BenchmarkCbor2ValUint64	20000000     92.0 ns/op	   8 B/op	 1 allocs/op
BenchmarkCbor2ValInt64	20000000     90.3 ns/op	   8 B/op	 1 allocs/op
BenchmarkCbor2ValFlt32	20000000     87.4 ns/op	   4 B/op	 1 allocs/op
BenchmarkCbor2ValFlt64	20000000     98.1 ns/op	   8 B/op	 1 allocs/op
BenchmarkCbor2ValBytes	 5000000    248 ns/op	  48 B/op	 2 allocs/op
BenchmarkCbor2ValText	10000000    217 ns/op	  32 B/op	 2 allocs/op
BenchmarkCbor2ValArr5	 1000000   1400 ns/op	 304 B/op	10 allocs/op
BenchmarkCbor2ValMap5	  500000   2850 ns/op	 496 B/op	18 allocs/op
BenchmarkCbor2ValTyp       50000  25685 ns/op   7783 B/op  140 allocs/op
```

**json to cbor**

* `null`, `true`, `false` json types are encodable into cbor
  format.
* `number` types are encoded based on configuration parameter
  `NumberKind`, which can be one of the following.
  * `JsonNumber` number is encoded as cbor-text (aka cbor-String)
    and the whole item is tagged as `tagJsonNumber` (tag-37).
  * `FloatNumber` number is encoded as cbor-float64.
  * `FloatNumber32` number is encoded as cbor-float32.
  * `IntNumber` number is encoded as cbor-int64.
  * `SmartNumber` if number is floating point then it is encoded
    as cbor-float64, else encoded as cbor-int64.
  * `SmartNumber32` if number is floating point then it is encoded
    as cbor-float32, else encoded as cbor-float32.
* `string` will be parsed and translated into utf8, and subsequently
  encoded as cbor-text. If config.JsonString() is set, string will be
  encoded simply as JSON-string and tagged as `tagJsonString` (tag-37).
* `arrays` can be encoded in `Stream` mode, using cbor's
  indefinite-length scheme, or in `LengthPrefix` mode.
* `properties` can be encoded in `Stream` mode, using cbor's
  indefinite-length scheme, or in `LengthPrefix` mode.

```text
BenchmarkJson2CborNull	50000000     29.6 ns/op  135.22 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborInt	20000000    112 ns/op     80.22 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborFlt	10000000    123 ns/op     81.12 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborBool	50000000     30.0 ns/op  166.90 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborStr	 5000000    239 ns/op    158.51 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborArr	 3000000    427 ns/op     70.19 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborMap	 1000000   1169 ns/op     53.00 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborTyp     200000  10136 ns/op    143.54 MB/s   0 B/op  0 allocs/op
```

**cbor to json**

* `nil`, `true`, `false` cbor types are transformed back to
  equivalent json types.
* `float32` and `float64` are transformed back to 32 bit
  JSON-float and 64 bit JSON-float respectively, in
  non-exponent format.
* `integers` are transformed back to JSON-integer representation,
  and integers exceeding 9223372036854775807 are not supported.
* `tagJsonNumber` are interpreted as it is into JSON number.
* `strings` are encoded into JSON-string using `encoding/json`
  package.
* `arrays` either with length prefix or with indefinite encoding
  are converted back to json array.
* `maps` either with length prefix or with indefinite encoding
  are converted back to json property.
* bytes-strings are not supported or transformed to json.
* cbor-text with indefinite encoding are not supported.
* simple type float16 are not supported.

```text
BenchmarkCbor2JsonNull  30000000    44.2 ns/op    22.63 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonInt   10000000   170 ns/op      52.69 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonFlt    5000000   256 ns/op      35.03 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonBool  30000000    44.3 ns/op    22.59 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonStr   10000000   217 ns/op     151.53 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonArr    5000000   401 ns/op      49.87 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonMap    1000000  1034 ns/op      51.24 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonTyp     200000  7081 ns/op     154.77 MB/s    0 B/op    0 allocs/op
```

**value to collate**

* `nil`, `true`, `false`, `float64`, `int64`, `int`, `Missing`,
  `string`, `[]byte`, `[]interface{}`, `map[string]interface{}`
  types are supported for collation.
* if configured as `JsonNumber`, `FloatNumber`, `FloatNumber32`
  number will be collated as floating point.
* if configured as `IntNumber` number will be collated as integer.
* if configured as `Decimal` number will be collated as
  small-decimal ( -1 >= num <= 1 ).
* if string value is MissingLiteral, it shall be collated as
  missing.

```text
BenchmarkVal2CollNil	200000000      8.30 ns/op    0 B/op	 0 allocs/op
BenchmarkVal2CollTrue	100000000     18.9 ns/op     0 B/op	 0 allocs/op
BenchmarkVal2CollFalse	100000000     19.8 ns/op     0 B/op	 0 allocs/op
BenchmarkVal2CollF64	 2000000     750 ns/op	     0 B/op	 0 allocs/op
BenchmarkVal2CollI64	 3000000     475 ns/op	     0 B/op	 0 allocs/op
BenchmarkVal2CollStr	30000000      47.2 ns/op     0 B/op	 0 allocs/op
BenchmarkVal2CollArr	 2000000     802 ns/op       0 B/op	 0 allocs/op
BenchmarkVal2CollMap	  500000    2688 ns/op     163 B/op	 9 allocs/op
```

**collate to value**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to value.

```text
BenchmarkColl2ValNil	200000000      9.26 ns/op    0 B/op   0 allocs/op
BenchmarkColl2ValTrue	50000000      38.0 ns/op     1 B/op   1 allocs/op
BenchmarkColl2ValFalse	50000000      38.4 ns/op     1 B/op   1 allocs/op
BenchmarkColl2ValF64	 5000000     348 ns/op       8 B/op   1 allocs/op
BenchmarkColl2ValI64	 5000000     247 ns/op       8 B/op   1 allocs/op
BenchmarkColl2ValMiss	20000000      88.6 ns/op    16 B/op   1 allocs/op
BenchmarkColl2ValStr	10000000     234 ns/op      32 B/op   2 allocs/op
BenchmarkColl2ValArr	 1000000    1070 ns/op     208 B/op   7 allocs/op
BenchmarkColl2ValMap	  500000    2704 ns/op     480 B/op  17 allocs/op
```

**json to collate**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are supported for collation.
* `number` is parsed as float64 and collated based on configuration:
  * if configured as `JsonNumber`, `FloatNumber`, `FloatNumber32`
    number will be collated as floating point.
  * if configured as `IntNumber` number will be collated as integer.
  * if configured as `Decimal` number will be collated as
    small-decimal ( -1 >= num <= 1 ).
* if string value is MissingLiteral, it shall be collated as
  missing.
* all other `string` value will be encoded into utf8 format before
  collating it.

```text
BenchmarkJson2CollNil	50000000     35.7 ns/op	  0 B/op    0 allocs/op
BenchmarkJson2CollTrue	50000000     35.5 ns/op	  0 B/op    0 allocs/op
BenchmarkJson2CollFalse	50000000     37.1 ns/op	  0 B/op    0 allocs/op
BenchmarkJson2CollF64	 1000000   1038 ns/op     8 B/op    1 allocs/op
BenchmarkJson2CollI64	 2000000    638 ns/op     8 B/op    1 allocs/op
BenchmarkJson2CollStr	 5000000    293 ns/op    32 B/op    1 allocs/op
BenchmarkJson2CollArr	 1000000   1562 ns/op    42 B/op    2 allocs/op
BenchmarkJson2CollMap	  500000   3550 ns/op   620 B/op   12 allocs/op
```

**collate to json**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are converted back to json.

```text
BenchmarkColl2JsonNil   100000000   16.3 ns/op     0 B/op    0 allocs/op
BenchmarkColl2JsonTrue  100000000   17.5 ns/op     0 B/op    0 allocs/op
BenchmarkColl2JsonFalse 100000000   18.1 ns/op     0 B/op    0 allocs/op
BenchmarkColl2JsonF64   10000000   123 ns/op       0 B/op    0 allocs/op
BenchmarkColl2JsonI64   20000000   104 ns/op       0 B/op    0 allocs/op
BenchmarkColl2JsonStr    2000000   660 ns/op      24 B/op    2 allocs/op
BenchmarkColl2JsonArr    2000000   903 ns/op      24 B/op    2 allocs/op
BenchmarkColl2JsonMap     300000  4395 ns/op     151 B/op   12 allocs/op
```

**cbor to collate**

* `null`, `true`, `false`, `float32`, `float64`, `integer`, `string`,
  `[]byte` (aka binary), `array`, `object` types in cbor can be
  collated.
* indefinite-length encoding for text and binary are not supported.
* LengthPrefix and Stream encoding for array and maps are supported.

```text
BenchmarkCbor2CollNil   30000000     40.4 ns/op    0 B/op  0 allocs/op
BenchmarkCbor2CollTrue  30000000     39.1 ns/op    0 B/op  0 allocs/op
BenchmarkCbor2CollFalse 50000000     38.5 ns/op    0 B/op  0 allocs/op
BenchmarkCbor2CollF64    2000000    844 ns/op      8 B/op  1 allocs/op
BenchmarkCbor2CollI64    3000000    523 ns/op      8 B/op  1 allocs/op
BenchmarkCbor2CollMiss  30000000     58.9 ns/op    0 B/op  0 allocs/op
BenchmarkCbor2CollStr   20000000     67.1 ns/op    0 B/op  0 allocs/op
BenchmarkCbor2CollArr    1000000   1209 ns/op      8 B/op  1 allocs/op
BenchmarkCbor2CollMap    1000000   2269 ns/op     49 B/op  3 allocs/op
```

**collate to cbor**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to cbor.

```text
BenchmarkColl2CborNil   100000000    20.3 ns/op    0 B/op    0 allocs/op
BenchmarkColl2CborTrue  100000000    20.7 ns/op    0 B/op    0 allocs/op
BenchmarkColl2CborFalse 100000000    21.4 ns/op    0 B/op    0 allocs/op
BenchmarkColl2CborF64    5000000    364 ns/op      8 B/op    1 allocs/op
BenchmarkColl2CborI64    5000000    268 ns/op      8 B/op    1 allocs/op
BenchmarkColl2CborMiss  30000000     43.6 ns/op    0 B/op    0 allocs/op
BenchmarkColl2CborStr   10000000    215 ns/op      0 B/op    0 allocs/op
BenchmarkColl2CborArr    2000000    620 ns/op     22 B/op    1 allocs/op
BenchmarkColl2CborMap    1000000   1838 ns/op     23 B/op    1 allocs/op
```

Notes
-----

* All supplied APIs will panic in case of error, applications can
  recover from panic, dump a stack trace along with input passed on to
  the API, and subsequently handle all such panics as a single valued
  error.
* maximum integer space shall be in int64.
* `Config` instances, and its APIs, are neither re-entrant not thread safe.

**list of changes from github.com/prataprc/collatejson**

* Codec type is renamed to Config.
* caller should make sure that the o/p buffer passed to encoding
  and decoding APIs are adequately sized.
* Name and signature of NewCodec() (now, NewConfig) has changed.
* configuration APIs,
  SortbyArrayLen, SortbyPropertyLen, UseMissing, NumberType all now return
  the config object back the caller - helps in call-chaining.
* all APIs panic instead of returning an error.
* output buffer should have its len() == cap(), so that encoder and decoder
  can avoid append and instead use buffer index.

License
-------

Copyright (c) 2015 Couchbase, Inc.
