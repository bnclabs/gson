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
* numbers can be interpreted as integer, or float64,
  *  `IntNumber` to interpret JSON number as integer whose size is
     defined by the platform.
  *  `FloatNumber` to interpret JSON number as 64-bit floating point.
* whitespace can be interpreted, based on configuration parameter
  `SpaceKind`, as `AnsiSpace` that should be faster
  than `UnicodeSpace`, while the later supports unicode whitespaces
  as well.

```text
BenchmarkJson2ValNil    30000000     38 ns/op   104 MB/s       0 B/op     0 allocs/op
BenchmarkJson2ValBool   20000000     66 ns/op    59 MB/s       1 B/op     1 allocs/op
BenchmarkJson2ValNum    10000000    164 ns/op    54 MB/s       8 B/op     1 allocs/op
BenchmarkJson2ValString  3000000    491 ns/op    65 MB/s      80 B/op     3 allocs/op
BenchmarkJson2ValArr5    1000000   1062 ns/op    28 MB/s     788 B/op     9 allocs/op
BenchmarkJson2ValMap5     500000   2654 ns/op    23 MB/s    5344 B/op    14 allocs/op
BenchmarkJson2ValTyp       50000  26077 ns/op    55 MB/s   23967 B/op   128 allocs/op
```

**value to json**

to convert value back to json text.

```
BenchmarkVal2JsonNil    100000000    14 ns/op  273 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonBool   50000000     25 ns/op  158 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonNum    10000000    212 ns/op   42 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonString 10000000    210 ns/op  181 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonArr5    5000000    257 ns/op  113 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonMap5    2000000    637 ns/op   80 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonTyp      200000   8437 ns/op  140 MB/s    0 B/op    0 allocs/op
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
  * `CborTagBytes` type supplied by cbor package, encoded with
    tag-24.
  * `regexp.Regexp` encoded with tag-35.
  * `CborTagPrefix` type supplied by cbor package, encoded
    with tag-55799.
* all other types shall cause a panic.

```text
BenchmarkVal2CborNull   50000000    45 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborTrue   30000000    40 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborUint64 20000000   168 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborFlt64  30000000   306 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborTBytes  5000000   365 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborText    5000000   402 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborArr5    2000000  1032 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborMap5    1000000  1332 ns/op     0 B/op  0 allocs/op
BenchmarkVal2CborTyp      200000  8833 ns/op     0 B/op  0 allocs/op
```


**cbor to value**

* reverse of all `value to cbor` encoding, described above, are
  supported.
* cannot decode `float16` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks shall be decoded outside
  this package using `IsIndefinite*()` and `IsBreakstop()` APIs.

```text
BenchmarkCbor2ValNull   50000000     40 ns/op      0 B/op    0 allocs/op
BenchmarkCbor2ValTrue   20000000     67 ns/op      1 B/op    1 allocs/op
BenchmarkCbor2ValUint64 20000000    100 ns/op      8 B/op    1 allocs/op
BenchmarkCbor2ValFlt64  20000000    116 ns/op      8 B/op    1 allocs/op
BenchmarkCbor2ValBytes   5000000    261 ns/op     48 B/op    2 allocs/op
BenchmarkCbor2ValText    5000000    224 ns/op     32 B/op    2 allocs/op
BenchmarkCbor2ValArr5    1000000   1379 ns/op    304 B/op    10 allocs/op
BenchmarkCbor2ValMap5     500000   3428 ns/op    496 B/op    18 allocs/op
BenchmarkCbor2ValTyp       50000  29681 ns/op   7784 B/op   140 allocs/op
```

**json to cbor**

* `null`, `true`, `false` json types are encodable into cbor
  format.
* `number` types are encoded based on configuration parameter
  `NumberKind`, which can be one of the following.
  * `FloatNumber` number is encoded as cbor-float64.
  * `FloatNumber32` number is encoded as cbor-float32.
  * `IntNumber` number is encoded as cbor-int64.
  * `SmartNumber` if number is floating point then it is encoded
    as cbor-float64, else encoded as cbor-int64.
  * `SmartNumber32` if number is floating point then it is encoded
    as cbor-float32, else encoded as cbor-float32.
* `string` will be parsed and translated into utf8, and subsequently
  encoded as cbor-text.
* `arrays` can be encoded in `Stream` mode, using cbor's
  indefinite-length scheme, or in `LengthPrefix` mode.
* `properties` can be encoded either using cbor's indefinite-length
   scheme (`Stream`), or using cbor's `LengthPrefix`.
* `property-keys` are always interpreted as string and encoded as 
   utf8 cbor-text.

```text
BenchmarkJson2CborNull  50000000     37 ns/op    107.93 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborBool  30000000     37 ns/op    133.19 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborInt   10000000    123 ns/op     72.76 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborFlt   10000000    140 ns/op     71.23 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborStr    5000000    255 ns/op    148.96 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborArr    3000000    459 ns/op     65.28 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborMap    1000000   1319 ns/op     46.99 MB/s   0 B/op  0 allocs/op
BenchmarkJson2CborTyp     200000  10928 ns/op    133.14 MB/s   0 B/op  0 allocs/op
```

**cbor to json**

* `nil`, `true`, `false` cbor types are transformed back to
  equivalent json types.
* `float32` and `float64` are transformed back to 32 bit
  JSON-float and 64 bit JSON-float respectively, in
  non-exponent format.
* `integers` are transformed back to JSON-integer representation,
  and integers exceeding 9223372036854775807 are not supported.
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
BenchmarkCbor2JsonNull  30000000    48 ns/op      20.43 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonBool  30000000    49 ns/op      20.39 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonInt   10000000   198 ns/op      45.37 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonFlt    5000000   281 ns/op      32.01 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonStr    5000000   240 ns/op     137.15 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonArr    3000000   431 ns/op      46.36 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonMap    1000000  1091 ns/op      48.57 MB/s    0 B/op    0 allocs/op
BenchmarkCbor2JsonTyp     200000  7898 ns/op     138.76 MB/s    0 B/op    0 allocs/op
```

**value to collate**

* `nil`, `true`, `false`, `float64`, `int64`, `int`, `Missing`,
  `string`, `[]byte`, `[]interface{}`, `map[string]interface{}`
  types are supported for collation.
* if configured as `FloatNumber`, `FloatNumber32` number will
  be collated as floating point.
* if configured as `IntNumber` number will be collated as integer.
* if configured as `Decimal` number will be collated as
  small-decimal ( -1 >= num <= 1 ).
* if string value is MissingLiteral, it shall be collated as
  missing.

```text
BenchmarkVal2CollNil    100000000      16 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollTrue   50000000       25 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollF64     2000000      871 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollI64     3000000      537 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollStr    20000000       65 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollArr     2000000      905 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollMap     1000000     1693 ns/op       0 B/op  0 allocs/op
BenchmarkVal2CollTyp      100000    17617 ns/op       0 B/op  0 allocs/op
```

**collate to value**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to value.

```text
BenchmarkColl2ValNil    100000000     14.5 ns/op      0 B/op    0 allocs/op
BenchmarkColl2ValTrue   30000000      44.7 ns/op      1 B/op    1 allocs/op
BenchmarkColl2ValF64     5000000     372 ns/op        8 B/op    1 allocs/op
BenchmarkColl2ValI64     5000000     279 ns/op        8 B/op    1 allocs/op
BenchmarkColl2ValStr    10000000     239 ns/op       32 B/op    2 allocs/op
BenchmarkColl2ValArr     1000000    1022 ns/op      208 B/op    7 allocs/op
BenchmarkColl2ValMap      500000    2630 ns/op      480 B/op   17 allocs/op
BenchmarkColl2ValTyp       50000   31729 ns/op     8005 B/op  133 allocs/op
```

**json to collate**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are supported for collation.
* `number` is parsed as float64 and collated based on configuration:
  * if configured as `FloatNumber`, `FloatNumber32` number will be
    collated as floating point.
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
BenchmarkJson2CollTyp      50000  27891 ns/op  5380 B/op   70 allocs/op
```

**collate to json**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are converted back to json.

```text
BenchmarkColl2JsonNil   100000000     16.3 ns/op     0 B/op    0 allocs/op
BenchmarkColl2JsonTrue  100000000     17.5 ns/op     0 B/op    0 allocs/op
BenchmarkColl2JsonFalse 100000000     18.1 ns/op     0 B/op    0 allocs/op
BenchmarkColl2JsonF64   10000000     123 ns/op       0 B/op    0 allocs/op
BenchmarkColl2JsonI64   20000000     104 ns/op       0 B/op    0 allocs/op
BenchmarkColl2JsonStr    2000000     660 ns/op      24 B/op    2 allocs/op
BenchmarkColl2JsonArr    2000000     903 ns/op      24 B/op    2 allocs/op
BenchmarkColl2JsonMap     300000    4395 ns/op     151 B/op   12 allocs/op
BenchmarkColl2JsonTyp      30000   41649 ns/op    1344 B/op  112 allocs/op
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
BenchmarkCbor2CollTyp     100000  14241 ns/op    367 B/op 11 allocs/op
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
BenchmarkColl2CborTyp     100000  16425 ns/op    167 B/op    5 allocs/op
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
* Name and signature of NewCodec() (now, NewDefaultConfig) has changed.
* configuration APIs,
  SortbyArrayLen, SortbyPropertyLen, UseMissing, NumberType all now return
  the config object back the caller - helps in call-chaining.
* all APIs panic instead of returning an error.
* output buffer should have its len() == cap(), so that encoder and decoder
  can avoid append and instead use buffer index.

License
-------

Copyright (c) 2015 Couchbase, Inc.
