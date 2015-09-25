What is what
------------

**json**

* Java Script Object Notation, also called JSON, RFC-7159.
* fast becoming the internet standard for data exchange.
* human readable format, not so friendly for machine representation.

**json-pointer**

* url like field locator within a json object, RFC-6901.
* make sense only for json arrays and objects, but to any level
  of nesting.
* json-pointers shall be unquoted before they are used for
  accessing into json text (or an equivalent representation),
  after unquoting segments within the pointer, each segment shall
  be binary compared with property keys.
* json-pointers can be used to access gson or cbor representation.
* documents encoded in cbor format using LengthPrefix are not
  supported by lookup APIs.

**value (aka gson)**

* golang object parsed from json, cbor or collate representation.
* json arrays are represeted in golang as `[]interface{}`.
* json objects, aka properties, are presented in golang as
  `map[string]interface{}`.
* gson objects support operations like, Get(), Set(), and
  Delete() on individual fields located by the json-pointer.

**cbor**

* Concise Binary Object Representation, also called CBOR, RFC-7049.
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

Transforms
----------

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
    BenchmarkScanNumFlt     10000000         139 ns/op    64.50 MB/s           8 B/op        1 allocs/op
    BenchmarkScanNumJsonNum 20000000         101 ns/op    88.66 MB/s          16 B/op        1 allocs/op
    BenchmarkScanString     10000000         136 ns/op   233.62 MB/s           0 B/op        0 allocs/op
    BenchmarkJson2ValArr5    1000000        1048 ns/op    28.61 MB/s         719 B/op        8 allocs/op
    BenchmarkJson2ValMap5     500000        3210 ns/op    19.31 MB/s        5170 B/op       14 allocs/op
    BenchmarkJson2ValTyp       50000       25399 ns/op    57.29 MB/s       17886 B/op      127 allocs/op
    BenchmarkJson2ValCgz          30    40470953 ns/op    47.95 MB/s    11606452 B/op   244295 allocs/op
```

as compared to using encoding/json for the same data sample:

```text
    BenchmarkUnmarshalFlt    1000000        1445 ns/op    6.23 MB/s        264 B/op        3 allocs/op
    BenchmarkUnmarshalNum    1000000        1473 ns/op    6.11 MB/s        264 B/op        3 allocs/op
    BenchmarkUnmarshalStr    1000000        2056 ns/op   15.56 MB/s        336 B/op        4 allocs/op
    BenchmarkUnmarshalArr5    300000        3706 ns/op    8.09 MB/s        320 B/op       10 allocs/op
    BenchmarkUnmarshalMap5    200000       10198 ns/op    6.08 MB/s        976 B/op       41 allocs/op
    BenchmarkUnmarshalTyp      20000       72231 ns/op   20.14 MB/s       6544 B/op      248 allocs/op
    BenchmarkUnmarshalCgz         20    71343632 ns/op   27.20 MB/s    8352375 B/op   284017 allocs/op
```

**value to json**

* to convert value back to json text golang's encoding/json package is
  used.
* `Encoder` interface{} is used to re-use o/p buffer.

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
    BenchmarkVal2CborNull   200000000    9.67 ns/op  0 B/op  0 allocs/op
    BenchmarkVal2CborTrue   100000000   20.6 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborFalse  100000000   21.3 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborUint64 50000000    27.8 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborInt64  50000000    28.0 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborFlt32  100000000   22.4 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborFlt64  50000000    26.9 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborBytes  50000000    39.4 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborText   30000000    45.8 ns/op   0 B/op  0 allocs/op
    BenchmarkVal2CborArr5   10000000   194 ns/op     0 B/op  0 allocs/op
    BenchmarkVal2CborMap5    3000000   455 ns/op     0 B/op  0 allocs/op
```


**cbor to value**

* reverse of all `value to cbor` encoding, described above, are
  supported.
* cannot decode `float16` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks shall be decoded outside
  this package using `IsIndefinite*()` and `IsBreakstop()` APIs.

```text
    BenchmarkCbor2ValNull   30000000     44.6 ns/op    0 B/op    0 allocs/op
    BenchmarkCbor2ValTrue   20000000     70.3 ns/op    1 B/op    1 allocs/op
    BenchmarkCbor2ValFalse  20000000     78.7 ns/op    1 B/op    1 allocs/op
    BenchmarkCbor2ValUint64 20000000     99.9 ns/op    8 B/op    1 allocs/op
    BenchmarkCbor2ValInt64  20000000     95.0 ns/op    8 B/op    1 allocs/op
    BenchmarkCbor2ValFlt32  20000000     93.5 ns/op    4 B/op    1 allocs/op
    BenchmarkCbor2ValFlt64  20000000    109 ns/op      8 B/op    1 allocs/op
    BenchmarkCbor2ValBytes   5000000    277 ns/op     48 B/op    2 allocs/op
    BenchmarkCbor2ValText   10000000    230 ns/op     32 B/op    2 allocs/op
    BenchmarkCbor2ValArr5    1000000   1535 ns/op    304 B/op   10 allocs/op
    BenchmarkCbor2ValMap5     500000   3071 ns/op    496 B/op   18 allocs/op
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
    BenchmarkJson2CborNull  50000000    32.3 ns/op  123.71 MB/s  0 B/op  0 allocs/op
    BenchmarkJson2CborInt   20000000   119 ns/op     75.55 MB/s  0 B/op  0 allocs/op
    BenchmarkJson2CborFlt   10000000   129 ns/op     77.27 MB/s  0 B/op  0 allocs/op
    BenchmarkJson2CborBool  50000000    32.8 ns/op  152.23 MB/s  0 B/op  0 allocs/op
    BenchmarkJson2CborStr    5000000   254 ns/op    149.15 MB/s  0 B/op  0 allocs/op
    BenchmarkJson2CborArr    3000000   469 ns/op     63.86 MB/s  0 B/op  0 allocs/op
    BenchmarkJson2CborMap    1000000  1276 ns/op     48.57 MB/s  0 B/op  0 allocs/op
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
    BenchmarkCbor2JsonNull  30000000    46.1 ns/op   21.69 MB/s     0 B/op    0 allocs/op
    BenchmarkCbor2JsonInt   10000000   194 ns/op     46.16 MB/s     0 B/op    0 allocs/op
    BenchmarkCbor2JsonFlt    5000000   279 ns/op     32.22 MB/s     0 B/op    0 allocs/op
    BenchmarkCbor2JsonBool  30000000    52.5 ns/op   19.04 MB/s     0 B/op    0 allocs/op
    BenchmarkCbor2JsonStr    2000000   846 ns/op     39.00 MB/s    24 B/op    2 allocs/op
    BenchmarkCbor2JsonArr    1000000  1148 ns/op     17.41 MB/s    24 B/op    2 allocs/op
    BenchmarkCbor2JsonMap     300000  5185 ns/op     10.22 MB/s   168 B/op   14 allocs/op
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
    BenchmarkGsonCollNil    200000000     11.6 ns/op     0 B/op  0 allocs/op
    BenchmarkGsonCollTrue   100000000     20.4 ns/op     0 B/op  0 allocs/op
    BenchmarkGsonCollFalse  100000000     21.7 ns/op     0 B/op  0 allocs/op
    BenchmarkGsonCollF64     2000000     815 ns/op       0 B/op  0 allocs/op
    BenchmarkGsonCollI64     3000000     510 ns/op       0 B/op  0 allocs/op
    BenchmarkGsonCollStr    30000000      54.7 ns/op     0 B/op  0 allocs/op
    BenchmarkGsonCollArr     2000000     809 ns/op       0 B/op  0 allocs/op
    BenchmarkGsonCollMap      500000    2929 ns/op     163 B/op  9 allocs/op
```

**collate to value**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to value.

```text
    BenchmarkCollGsonNil    100000000    11.4 ns/op      0 B/op   0 allocs/op
    BenchmarkCollGsonTrue    50000000     39.8 ns/op     1 B/op   1 allocs/op
    BenchmarkCollGsonFalse   50000000     40.0 ns/op     1 B/op   1 allocs/op
    BenchmarkCollGsonF64      5000000    376 ns/op       8 B/op   1 allocs/op
    BenchmarkCollGsonI64      5000000    275 ns/op       8 B/op   1 allocs/op
    BenchmarkCollGsonStr      5000000    254 ns/op      32 B/op   2 allocs/op
    BenchmarkCollGsonArr      1000000   1147 ns/op     208 B/op   7 allocs/op
    BenchmarkCollGsonMap       500000   3060 ns/op     480 B/op  17 allocs/op
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
    BenchmarkJsonCollNil    50000000     37.9 ns/op   0 B/op    0 allocs/op
    BenchmarkJsonCollTrue   50000000     38.3 ns/op   0 B/op    0 allocs/op
    BenchmarkJsonCollFalse  30000000     39.0 ns/op   0 B/op    0 allocs/op
    BenchmarkJsonCollF64     1000000   1074 ns/op     8 B/op    1 allocs/op
    BenchmarkJsonCollI64     2000000    695 ns/op     8 B/op    1 allocs/op
    BenchmarkJsonCollStr     5000000    311 ns/op    32 B/op    1 allocs/op
    BenchmarkJsonCollArr     1000000   1629 ns/op    42 B/op    2 allocs/op
    BenchmarkJsonCollMap      300000   3912 ns/op   620 B/op   12 allocs/op
```

**cbor to collate**

* `null`, `true`, `false`, `float32`, `float64`, `integer`, `string`,
  `[]byte` (aka binary), `array`, `object` types in cbor can be
  collated.
* indefinite-length encoding for text and binary are not supported.
* LengthPrefix and Stream encoding for array and maps are supported.

**collate to cbor**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to cbor.

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
