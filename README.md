[![Build Status](https://travis-ci.org/prataprc/gson.png)](https://travis-ci.org/prataprc/gson)
[![Coverage Status](https://coveralls.io/repos/prataprc/gson/badge.png?branch=master&service=github)](https://coveralls.io/github/prataprc/gson?branch=master)
[![GoDoc](https://godoc.org/github.com/prataprc/gson?status.png)](https://godoc.org/github.com/prataprc/gson)

Topics
======

* [what is what](#what-is-what)
* [transforms](#transforms)
* [getting started](docs/gettingstarted.md)
* [notes](#notes)

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
BenchmarkJson2ValNil      50000000     31 ns/op  129.17 MB/s      0 B/op    0 allocs/op
BenchmarkJson2ValBool     30000000     57 ns/op   70.03 MB/s      1 B/op    1 allocs/op
BenchmarkJson2ValNum      10000000    136 ns/op   66.08 MB/s      8 B/op    1 allocs/op
BenchmarkJson2ValString    3000000    480 ns/op   66.65 MB/s     84 B/op    3 allocs/op
BenchmarkJson2ValArr5      1000000   1049 ns/op   28.58 MB/s    393 B/op    9 allocs/op
BenchmarkJson2ValMap5      1000000   1912 ns/op   32.41 MB/s    690 B/op   14 allocs/op
BenchmarkJson2ValTyp         50000  27432 ns/op   53.04 MB/s  19528 B/op  128 allocs/op
```

**value to json**

to convert value back to json text.

```
BenchmarkVal2JsonNil     100000000    148 ns/op   270.85 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonBool     50000000     25 ns/op   154.84 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonNum      10000000    168 ns/op    53.26 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonString   10000000    167 ns/op   227.38 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonArr5      5000000    262 ns/op   110.33 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonMap5      3000000    537 ns/op    94.91 MB/s    0 B/op    0 allocs/op
BenchmarkVal2JsonTyp        200000   6713 ns/op   176.65 MB/s    0 B/op    0 allocs/op
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
BenchmarkVal2CborNull      100000000    14 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborTrue      50000000     25 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborUint64-8  50000000     32 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborFlt64     50000000     29 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborTBytes    30000000     45 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborText      30000000     51 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborArr5      10000000    198 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborMap5       3000000    449 ns/op    0 B/op    0 allocs/op
BenchmarkVal2CborTyp         500000   3586 ns/op    0 B/op    0 allocs/op

```


**cbor to value**

* reverse of all `value to cbor` encoding, described above, are
  supported.
* cannot decode `float16` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks shall be decoded outside
  this package using `IsIndefinite*()` and `IsBreakstop()` APIs.

```text
BenchmarkCbor2ValNull    50000000     38 ns/op    0 B/op    0 allocs/op
BenchmarkCbor2ValTrue    20000000     64 ns/op    1 B/op    1 allocs/op
BenchmarkCbor2ValUint64  20000000     69 ns/op    8 B/op    1 allocs/op
BenchmarkCbor2ValFlt64   20000000     77 ns/op    8 B/op    1 allocs/op
BenchmarkCbor2ValBytes   10000000    170 ns/op   48 B/op    2 allocs/op
BenchmarkCbor2ValText    10000000    163 ns/op   32 B/op    2 allocs/op
BenchmarkCbor2ValArr5     1000000   1047 ns/op  304 B/op   10 allocs/op
BenchmarkCbor2ValMap5     1000000   2020 ns/op  496 B/op   18 allocs/op
BenchmarkCbor2ValTyp       100000  19255 ns/op  784 B/op  140 allocs/op
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
BenchmarkJson2CborNull  50000000    31 ns/op 127 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborBool  50000000    31 ns/op 160 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborInt   20000000   102 ns/op  87 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborFlt   20000000   111 ns/op  89 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborStr   10000000   206 ns/op 183 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborArr    5000000   371 ns/op  80 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborMap    2000000   982 ns/op  63 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborTyp     200000  9273 ns/op 156 MB/s  0 B/op  0 allocs/op
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
BenchmarkCbor2JsonNull  30000000    50 ns/op   19 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonBool  30000000    48 ns/op   20 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonInt   10000000   140 ns/op   64 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonFlt   10000000   239 ns/op   37 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonStr   10000000   187 ns/op  175 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonArr    5000000   392 ns/op   50 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonMap    1000000  1003 ns/op   52 MB/s  0 B/op   0 allocs/op
BenchmarkCbor2JsonTyp     200000  6871 ns/op  159 MB/s  0 B/op   0 allocs/op
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
BenchmarkVal2CollNil    100000000     13 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollTrue    50000000     25 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollF64      2000000    692 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollI64      3000000    405 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollStr     30000000     58 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollArr      2000000    707 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollMap      1000000   1458 ns/op   0 B/op  0 allocs/op
BenchmarkVal2CollTyp       100000  13511 ns/op   0 B/op  0 allocs/op
```

**collate to value**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to value.

```text
BenchmarkColl2ValNil   100000000    10 ns/op     0 B/op    0 allocs/op
BenchmarkColl2ValTrue  50000000     35 ns/op     1 B/op    1 allocs/op
BenchmarkColl2ValF64    3000000    499 ns/op   144 B/op    4 allocs/op
BenchmarkColl2ValI64    5000000    396 ns/op   144 B/op    4 allocs/op
BenchmarkColl2ValStr   10000000    183 ns/op    32 B/op    2 allocs/op
BenchmarkColl2ValArr    2000000    968 ns/op   336 B/op   10 allocs/op
BenchmarkColl2ValMap    1000000   2222 ns/op   616 B/op   20 allocs/op
BenchmarkColl2ValTyp     100000  22437 ns/op  8648 B/op  145 allocs/op
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
BenchmarkJson2CollNil   50000000     34 ns/op   0 B/op   0 allocs/op
BenchmarkJson2CollTrue  50000000     35 ns/op   0 B/op   0 allocs/op
BenchmarkJson2CollF64    2000000    850 ns/op   0 B/op   0 allocs/op
BenchmarkJson2CollI64    3000000    469 ns/op   0 B/op   0 allocs/op
BenchmarkJson2CollStr   10000000    223 ns/op   0 B/op   0 allocs/op
BenchmarkJson2CollArr    1000000   1238 ns/op   1 B/op   0 allocs/op
BenchmarkJson2CollMap    1000000   2179 ns/op   2 B/op   0 allocs/op
BenchmarkJson2CollTyp     100000  22750 ns/op   3 B/op   0 allocs/op

```

**collate to json**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are converted back to json.

```text
BenchmarkColl2JsonNil   100000000    19 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonTrue  100000000    19 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonF64   10000000    239 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonI64   10000000    228 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonStr   10000000    227 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonArr    3000000    568 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonMap    1000000   1774 ns/op    0 B/op   0 allocs/op
BenchmarkColl2JsonTyp     100000  17297 ns/op    0 B/op   0 allocs/op
```

**cbor to collate**

* `null`, `true`, `false`, `float32`, `float64`, `integer`, `string`,
  `[]byte` (aka binary), `array`, `object` types in cbor can be
  collated.
* indefinite-length encoding for text and binary are not supported.
* LengthPrefix and Stream encoding for array and maps are supported.

```text
BenchmarkCbor2CollNil   30000000      44 ns/op   0 B/op  0 allocs/op
BenchmarkCbor2CollTrue  30000000      45 ns/op   0 B/op  0 allocs/op
BenchmarkCbor2CollF64    2000000     666 ns/op   0 B/op  0 allocs/op
BenchmarkCbor2CollI64    3000000     413 ns/op   0 B/op  0 allocs/op
BenchmarkCbor2CollStr   20000000      68 ns/op   0 B/op  0 allocs/op
BenchmarkCbor2CollArr    1000000    1056 ns/op   0 B/op  0 allocs/op
BenchmarkCbor2CollMap    1000000    1805 ns/op   1 B/op  0 allocs/op
BenchmarkCbor2CollTyp     100000   15447 ns/op  33 B/op  0 allocs/op

```

**collate to cbor**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to cbor.

```text
BenchmarkColl2CborNil   100000000     17 ns/op    0 B/op  0 allocs/op
BenchmarkColl2CborTrue  100000000     17 ns/op    0 B/op  0 allocs/op
BenchmarkColl2CborF64    10000000    177 ns/op    0 B/op  0 allocs/op
BenchmarkColl2CborI64    10000000    166 ns/op    0 B/op  0 allocs/op
BenchmarkColl2CborStr    10000000    210 ns/op    0 B/op  0 allocs/op
BenchmarkColl2CborArr     3000000    481 ns/op    0 B/op  0 allocs/op
BenchmarkColl2CborMap     1000000   1675 ns/op    2 B/op  0 allocs/op
BenchmarkColl2CborTyp      100000  14260 ns/op   54 B/op  0 allocs/op
```

Notes
-----

* All supplied APIs will panic in case of error, applications can
  recover from panic, dump a stack trace along with input passed on to
  the API, and subsequently handle all such panics as a single valued
  error.
* maximum integer space shall be in int64.
* `Config` instances, and its APIs, are neither re-entrant not thread safe.
* encoding/json.Number is not supported yet.

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
