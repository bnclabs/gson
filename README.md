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
BenchmarkJson2ValNil    10000000         207 ns/op   19.28 MB/s            0 B/op       0 allocs/op
BenchmarkJson2ValBool    5000000         262 ns/op   15.25 MB/s            1 B/op       1 allocs/op
BenchmarkJson2ValNum     3000000         582 ns/op   15.46 MB/s            8 B/op       1 allocs/op
BenchmarkJson2ValString     2000      632553 ns/op    0.05 MB/s      1048730 B/op       8 allocs/op
BenchmarkJson2ValArr5       2000      634392 ns/op    0.05 MB/s      1048938 B/op      15 allocs/op
BenchmarkJson2ValMap5       1000     1329876 ns/op    0.05 MB/s      2097768 B/op      28 allocs/op
BenchmarkJson2ValTyp         100    17402659 ns/op    0.08 MB/s     28321828 B/op     290 allocs/op
```

**value to json**

to convert value back to json text.

```
BenchmarkVal2JsonNil     10000000    198 ns/op  20.19 MB/s  0 B/op  0 allocs/op
BenchmarkVal2JsonBool    10000000    216 ns/op  18.48 MB/s  0 B/op  0 allocs/op
BenchmarkVal2JsonNum       500000   2376 ns/op   3.79 MB/s  0 B/op  0 allocs/op
BenchmarkVal2JsonString   1000000   1349 ns/op  28.15 MB/s  0 B/op  0 allocs/op
BenchmarkVal2JsonArr5     1000000   2167 ns/op  13.38 MB/s  0 B/op  0 allocs/op
BenchmarkVal2JsonMap5      300000   4394 ns/op  11.61 MB/s  0 B/op  0 allocs/op
BenchmarkVal2JsonTyp        20000  66023 ns/op  17.96 MB/s  0 B/op  0 allocs/op
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
BenchmarkVal2CborNull    10000000    155 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborTrue    10000000    170 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborUint64   5000000    365 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborFlt64    5000000    368 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborTBytes   3000000    471 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborText     3000000    585 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborArr5     1000000   1330 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborMap5      500000   3550 ns/op  0 B/op  0 allocs/op
BenchmarkVal2CborTyp        50000  38530 ns/op  0 B/op  0 allocs/op

```


**cbor to value**

* reverse of all `value to cbor` encoding, described above, are
  supported.
* cannot decode `float16` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks shall be decoded outside
  this package using `IsIndefinite*()` and `IsBreakstop()` APIs.

```text
BenchmarkCbor2ValNull    10000000    214 ns/op     0 B/op   0 allocs/op
BenchmarkCbor2ValTrue     5000000    262 ns/op     1 B/op   1 allocs/op
BenchmarkCbor2ValUint64   3000000    451 ns/op     8 B/op   1 allocs/op
BenchmarkCbor2ValFlt64    3000000    452 ns/op     8 B/op   1 allocs/op
BenchmarkCbor2ValBytes    2000000    639 ns/op    48 B/op   2 allocs/op
BenchmarkCbor2ValText     2000000    677 ns/op    32 B/op   2 allocs/op
BenchmarkCbor2ValArr5      500000   3363 ns/op   304 B/op  10 allocs/op
BenchmarkCbor2ValMap5      200000   6540 ns/op   496 B/op  18 allocs/op
BenchmarkCbor2ValTyp        20000  61061 ns/op  7786 B/op 140 allocs/op
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
BenchmarkJson2CborNull  5000000    310 ns/op  12.89 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborBool  5000000    309 ns/op  16.15 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborInt   2000000    974 ns/op   9.24 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborFlt   2000000    869 ns/op  11.50 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborStr   1000000   1791 ns/op  21.21 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborArr    500000   2659 ns/op  11.28 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborMap    200000   7895 ns/op   7.85 MB/s  0 B/op  0 allocs/op
BenchmarkJson2CborTyp     20000  77367 ns/op  18.81 MB/s  0 B/op  0 allocs/op
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
BenchmarkCbor2JsonNull  3000000    394 ns/op   2.54 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonBool  3000000    414 ns/op   2.41 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonInt   1000000   1643 ns/op   5.48 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonFlt    500000   2732 ns/op   3.29 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonStr    500000   2662 ns/op  12.39 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonArr    500000   3592 ns/op   5.57 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonMap    200000   9336 ns/op   5.68 MB/s  0 B/op  0 allocs/op
BenchmarkCbor2JsonTyp     20000  72542 ns/op  15.11 MB/s  0 B/op  0 allocs/op
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
BenchmarkVal2CollNil    10000000     202 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollTrue   10000000     198 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollFalse  10000000     199 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollF64      200000    6887 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollI64      300000    3862 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollStr     2000000     772 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollArr      200000    6708 ns/op      0 B/op   0 allocs/op
BenchmarkVal2CollMap       50000   26696 ns/op  16416 B/op   2 allocs/op
BenchmarkVal2CollTyp       10000  174548 ns/op  49250 B/op   6 allocs/op
```

**collate to value**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to value.

```text
BenchmarkColl2ValNil   10000000     128 ns/op     0 B/op    0 allocs/op
BenchmarkColl2ValTrue  10000000     190 ns/op     1 B/op    1 allocs/op
BenchmarkColl2ValF64     300000    4355 ns/op   144 B/op    4 allocs/op
BenchmarkColl2ValI64     500000    3129 ns/op   144 B/op    4 allocs/op
BenchmarkColl2ValStr    1000000    1428 ns/op    32 B/op    2 allocs/op
BenchmarkColl2ValArr     200000    6248 ns/op   336 B/op   10 allocs/op
BenchmarkColl2ValMap     100000   12240 ns/op   616 B/op   20 allocs/op
BenchmarkColl2ValTyp      10000  148687 ns/op  8648 B/op  145 allocs/op
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
BenchmarkJson2CollNil   5000000       332 ns/op         0 B/op    0 allocs/op
BenchmarkJson2CollTrue  5000000       323 ns/op         0 B/op    0 allocs/op
BenchmarkJson2CollF64    200000      8057 ns/op         0 B/op    0 allocs/op
BenchmarkJson2CollI64    300000      4368 ns/op         0 B/op    0 allocs/op
BenchmarkJson2CollStr      2000    635015 ns/op   1048647 B/op    5 allocs/op
BenchmarkJson2CollArr      2000    645310 ns/op   1048645 B/op    5 allocs/op
BenchmarkJson2CollMap      1000   1372242 ns/op   2138328 B/op   16 allocs/op
BenchmarkJson2CollTyp       100  17937836 ns/op  28436440 B/op  163 allocs/op
```

**collate to json**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are converted back to json.

```text
BenchmarkColl2JsonNil    5000000       254 ns/op        0 B/op    0 allocs/op
BenchmarkColl2JsonTrue   5000000       260 ns/op        0 B/op    0 allocs/op
BenchmarkColl2JsonF64     200000      8044 ns/op      129 B/op    3 allocs/op
BenchmarkColl2JsonI64     300000      5108 ns/op      129 B/op    3 allocs/op
BenchmarkColl2JsonStr       2000    628639 ns/op  1048647 B/op    5 allocs/op
BenchmarkColl2JsonArr       2000    642285 ns/op  1048782 B/op    8 allocs/op
BenchmarkColl2JsonMap        500   3796742 ns/op  6292014 B/op   37 allocs/op
BenchmarkColl2JsonTyp         50  35517962 ns/op 58724887 B/op  333 allocs/op
```

**cbor to collate**

* `null`, `true`, `false`, `float32`, `float64`, `integer`, `string`,
  `[]byte` (aka binary), `array`, `object` types in cbor can be
  collated.
* indefinite-length encoding for text and binary are not supported.
* LengthPrefix and Stream encoding for array and maps are supported.

```text
BenchmarkCbor2CollNil   5000000      325 ns/op        0 B/op   0 allocs/op
BenchmarkCbor2CollTrue  5000000      313 ns/op        0 B/op   0 allocs/op
BenchmarkCbor2CollF64    200000     7246 ns/op        0 B/op   0 allocs/op
BenchmarkCbor2CollStr   1000000     1060 ns/op        0 B/op   0 allocs/op
BenchmarkCbor2CollArr    200000    10238 ns/op        0 B/op   0 allocs/op
BenchmarkCbor2CollMap      2000   787087 ns/op  1089654 B/op   8 allocs/op
BenchmarkCbor2CollTyp       500  2500523 ns/op  3269067 B/op  25 allocs/op
```

**collate to cbor**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to cbor.

```text
BenchmarkColl2CborNil   10000000       209 ns/op         0 B/op    0 allocs/op
BenchmarkColl2CborTrue  10000000       218 ns/op         0 B/op    0 allocs/op
BenchmarkColl2CborF64     300000      4692 ns/op       129 B/op    3 allocs/op
BenchmarkColl2CborStr       2000    625580 ns/op   1049174 B/op    5 allocs/op
BenchmarkColl2CborArr       2000    632711 ns/op   1049308 B/op    8 allocs/op
BenchmarkColl2CborMap        500   3814868 ns/op   6296307 B/op   37 allocs/op
BenchmarkColl2CborTyp         30  36413341 ns/op  59673817 B/op  337 allocs/op
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
