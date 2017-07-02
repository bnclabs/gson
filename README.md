Object formats and notations
============================

[![Build Status](https://travis-ci.org/prataprc/gson.png)](https://travis-ci.org/prataprc/gson)
[![Coverage Status](https://coveralls.io/repos/prataprc/gson/badge.png?branch=master&service=github)](https://coveralls.io/github/prataprc/gson?branch=master)
[![GoDoc](https://godoc.org/github.com/prataprc/gson?status.png)](https://godoc.org/github.com/prataprc/gson)
[![Go Report Card](https://goreportcard.com/badge/github.com/prataprc/gson)](https://goreportcard.com/report/github.com/prataprc/gson)
[![GitPitch](https://gitpitch.com/assets/badge.svg)](https://gitpitch.com/prataprc/gson/master?grs=github&t=white)

**Goals**:

- High performance algorithms for data transformation, serialization and
  manipulation.
- Based on well established standards.
- ZERO allocation when transforming from one format to another, except
  for APIs creating golang values from encoded data.
- [JSON](http://json.org) for web.
- [CBOR](http://cbor.io) for machine.
- [Binary-Collation](docs/collate.md) for crazy fast comparison/sorting.

**This package is under continuous development, but the APIs are fairly stable**.

Quick Links
-----------

* [slides on gson](https://gitpitch.com/prataprc/gson/master?grs=github&t=white)
* [What is what](#what-is-what)
* [Configuration](#configuration)
* [Transforms](#transforms)
* [Understanding collation](docs/collate.md)
* [Getting started](docs/gettingstarted.md)
* [Play with command line](cmd/README.md)
* [Articles related to gson](#articles)
* [How to contribute](#how-to-contribute)

What is what
------------

**JSON**

* Java Script Object Notation, also called [JSON](http://www.json.org/),
  [RFC-7159](https://tools.ietf.org/html/rfc7159).
* Fast becoming the internet standard for data exchange.
* Human readable format, not so friendly for machine representation.

**Value (aka gson)**

* Golang object parsed from JSON, CBOR or collate representation.
* JSON arrays are represented in golang as `[]interface{}`.
* JSON objects, aka properties, are presented in golang as
  `map[string]interface{}`.
* Following golang-types can be transformed to JSON, CBOR, or,
  [Binary-collation](docs/collate.md): `nil`, `bool`,
  `byte, int8, int16, uint16, int32, uint32, int, uint, int64, uint64`,
  `float32, float64`,
  `string`, `[]interface{}`, `map[string]interface{}`,
  `[][2]interface{}`.
* For type `[][2]interface{}`, first item is treated as key (string) and
  second item is treated as value, hence equivalent to
  `map[string]interface{}`.
* Gson objects support operations like, Get(), Set(), and
  Delete() on individual fields located by the JSON-POINTER.

**CBOR**

* Concise Binary Object Representation, also called [CBOR](http://cbor.io/),
  [RFC-7049](https://tools.ietf.org/html/rfc7049).
* Machine friendly, designed for IoT, inter-networking of light weight
  devices, and easy to implement in many languages.
* Can be used for more than data exchange, left to user
  imagination :) ...

**Binary-Collation**

* A custom encoding based on a [paper](docs/collate.pdf) and improvised to
  handle JSON specification.
* Binary representation preserving the sort order.
* Transform back to original JSON from binary representation.
* Numbers can be encoded in three ways - as integers, or as
  small-decimals, or as floating-point represented in exponent
  form.
* More details can be found [here](docs/collate.md)

**JSON-Pointer**

* URL like field locator within a JSON object,
  [RFC-6901](https://tools.ietf.org/html/rfc6901).
* Make sense only for JSON arrays and objects, but to any level
  of nesting.
* JSON-pointers shall be unquoted before they are used for
  accessing into JSON text (or an equivalent representation),
  after unquoting segments within the pointer, each segment shall
  be binary compared with property keys.
* Documents encoded in CBOR format using LengthPrefix are not
  supported by lookup APIs.

**Performance and memory pressure**

Following Benchmark is made on a map data which has a shape similar to:

```go
{"key1": nil, "key2": true, "key3": false,
"key4": "hello world", "key5": 10.23122312}
```
or,
```go
{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}
```

```text
BenchmarkVal2JsonMap5-8  3000000   461 ns/op    0 B/op  0 allocs/op
BenchmarkVal2CborMap5-8  5000000   262 ns/op    0 B/op  0 allocs/op
BenchmarkVal2CollMap-8   1000000  1321 ns/op  128 B/op  2 allocs/op
BenchmarkJson2CborMap-8  2000000   838 ns/op    0 B/op  0 allocs/op
BenchmarkCbor2JsonMap-8  2000000  1010 ns/op    0 B/op  0 allocs/op
BenchmarkJson2CollMap-8  1000000  1825 ns/op  202 B/op  2 allocs/op
BenchmarkColl2JsonMap-8  1000000  2028 ns/op  434 B/op  6 allocs/op
BenchmarkCbor2CollMap-8  1000000  1692 ns/op  131 B/op  2 allocs/op
BenchmarkColl2CborMap-8  1000000  1769 ns/op  440 B/op  6 allocs/op
```

Though converting to golang value incurs cost.

```text
BenchmarkJson2ValMap5    1000000  1621 ns/op   699 B/op  14 allocs/op
BenchmarkCbor2ValMap5    1000000  1711 ns/op   496 B/op  18 allocs/op
BenchmarkColl2ValMap     1000000  2235 ns/op  1440 B/op  33 allocs/op
```

Configuration
-------------

**Configuration APIs are not re-entrant**. For concurrent use of Gson, please
create a `gson.Config{}` per routine, or protect them with a mutex.

**NumberKind**

There are two ways to treat numbers in Gson, as integers (upto 64-bit width)
or floating-point (float64).

- **FloatNumber** configuration can be used to tell Gson to treat all numbers
as floating point. This has the convenience of having precision between
discrete values, but suffers round of errors and inability to represent
integer values greater than 2^53. **DEFAULT choice**.

- **SmartNumber** will use int64, uint64 for representing integer values and
use float64 when decimal precision is required. Choosing this option, Gson
might need more CPU time.

Can be configured per configuration instance via `SetNumberKind()`.

**MaxKeys**

Maximum number of keys allowed in a property object. This can be configured
globally via `gson.MaxKeys` or per configuration via `SetMaxkeys()`.

**Memory-pools**

Gson uses memory pools:

* Pool of Byte blocks for listing json-pointers.
* Pool of Byte blocks for for encoding / decoding string types,
  and property keys.
* Pool of set of strings for sorting keys within property.

Memory foot print of gson depends on the pools size, maximum length of
input string, number keys in input property-map.

Mempools can be configured globally via
`gson.MaxStringLen, gson.MaxKeys, gson.MaxCollateLen, gson.MaxJsonpointerLen`
package variables or per configuration instance via `ResetPools()`.

**CBOR ContainerEncoding**

In CBOR both map and array (called container types) can be encoded as length
followed by items, or items followed by end-marker.

- **LengthPrefix** to encode length of container type first, followed by
  each item in the container.
- **Stream** to encode each item in the container as it appears in the input
  stream, finally ending it with a End-Stream-Marker. **DEFAULT choice**.

Can be configured per configuration instance via `SetContainerEncoding`

**JSON Strict**

- If configured as true, encode / decode JSON strings operations will use
Golang's encoding / JSON package.

Can be configured per configuration instance via `SetStrict`.

**JSON SpaceKind**

How to interpret space characters ? There are two options:

- **AnsiSpace** will be faster but does not support unicode.
- **UnicodeSpace** will be slower but supports unicode. **DEFAULT choice**.

Can be configured per configuration instance via `SetSpaceKind`.

**Collate ArrayLenPrefix**

While sorting array, which is a container type, should collation algorithm
consider the arity of the array ? If ArrayLenPrefix prefix is configured as
true, arrays with more number of items will sort after arrays with lesser
number of items.

Can be configured per configuration instance via `SortbyArrayLen`.

**Collate PropertyLenPrefix**

While sorting property-map, which is a container type, should collation
algorithm consider number of entries in the map ? If PropertyLenPrefix is
configured as true, maps with more number of items will sort after maps
with lesser number of items.

Can be configured per configuration instance via `SortbyPropertyLen`.

**JSON-Pointer JsonpointerLength**

Maximum length a JSON-pointer string can take. Can be configured globally
via MaxJsonpointerLen or per configuration instance via `SetJptrlen`.

NOTE: JSON pointers are composed of path segments, there is an upper limit
to the number of path-segments a JSON pointer can have. If your configuration
exceeds that limit, try increasing the JsonpointerLength.

Transforms
----------

![transforms](docs/transforms.png)

**JSON to value**

* Gson uses custom parser that must be faster than encoding/JSON.
* Numbers can be interpreted as integer, or float64,
  -  `FloatNumber` to interpret JSON number as 64-bit floating point.
  -  `SmartNumber` to interpret JSON number as int64, or uint64, or float64.
* Whitespace can be interpreted, based on configuration type `SpaceKind`.
  SpaceKind can be one of the following `AnsiSpace` or `UnicodeSpace`.
  - `AnsiSpace` that should be faster
  - `UnicodeSpace` supports unicode white-spaces as well.

**value to CBOR**

* Golang types `nil`, `true`, `false` are encodable into CBOR
  format.
* All Golang `number` types, including signed, unsigned, and
  floating-point variants, are encodable into CBOR format.
* Type `[]byte` is encoded as CBOR byte-string.
* Type `string` is encoded as CBOR text.
* Generic `array` is interpreted as Golang `[]interface{}` and
  encoded as CBOR array.
  * With `LengthPrefix` option for ContainerEncoding, arrays and
    maps are encoded with its length.
  * With `Stream` option, arrays and maps are encoded using
    Indefinite and Breakstop encoding.
* Generic `property` is interpreted as golang `[][2]interface{}`
  and encoded as CBOR array of 2-element array, where the first item
  is key represented as string and second item is any valid JSON
  value.
* Before encoding `map[string]interface{}` type, use
  `GolangMap2cborMap()` function to transform them to
  `[][2]interface{}`.
* Following golang data types are encoded using CBOR-tags,
  * Type `time.Time` encoded with tag-0.
  * Type `Epoch` type supplied by CBOR package, encoded
    with tag-1.
  * Type `EpochMicro` type supplied by CBOR package, encoded
    with tag-1.
  * Type `math/big.Int` positive numbers are encoded with tag-2, and
    negative numbers are encoded with tag-3.
  * Type `DecimalFraction` type supplied by CBOR package,
    encoded with tag-4.
  * Type `BigFloat` type supplied by CBOR package, encoded
    with tag-5.
  * Type `CborTagBytes` type supplied by CBOR package, encoded with
    tag-24.
  * Type `regexp.Regexp` encoded with tag-35.
  * Type `CborTagPrefix` type supplied by CBOR package, encoded
    with tag-55799.
* All other types shall cause a panic.

**CBOR to value**

* Reverse of all `value to CBOR` encoding, described above, are
  supported.
* Cannot decode `float16` type and int64 > 9223372036854775807.
* Indefinite byte-string chunks, text chunks shall be decoded outside
  this package using `IsIndefinite*()` and `IsBreakstop()` APIs.

**JSON to CBOR**

* JSON Types `null`, `true`, `false` are encodable into CBOR format.
* Types `number` are encoded based on configuration type `NumberKind`,
  which can be one of the following.
  * If config.nk is FloatNumber, all numbers are encoded as CBOR-float64.
  * If config.nk is SmartNumber, all JSON float64 numbers are encoded as
    CBOR-float64, and, all JSON positive integers are encoded as
    CBOR-uint64, and, all JSON negative integers are encoded as
    CBOR-int64.
* Type `string` will be parsed and translated into UTF-8, and subsequently
  encoded as CBOR-text.
* Type `arrays` can be encoded in `Stream` mode, using CBOR's
  indefinite-length scheme, or in `LengthPrefix` mode.
* Type `properties` can be encoded either using CBOR's indefinite-length
  scheme (`Stream`), or using CBOR's `LengthPrefix`.
* Property-keys are always interpreted as string and encoded as 
  UTF-8 CBOR-text.

**CBOR to JSON**

* CBOR types `nil`, `true`, `false` are transformed back to equivalent
  JSON types.
* Types `float32` and `float64` are transformed back to 32 bit
  JSON-float and 64 bit JSON-float respectively, in
  non-exponent format.
* Type `integer` is transformed back to JSON-integer representation,
  and integers exceeding 9223372036854775807 are not supported.
* Type `array` either with length prefix or with indefinite encoding
  are converted back to JSON array.
* Type `map` either with length prefix or with indefinite encoding
  are converted back to JSON property.
* Type bytes-strings are not supported or transformed to JSON.
* Type CBOR-text with indefinite encoding are not supported.
* Type Simple type float16 are not supported.

For transforming to and from binary-collation refer [here](docs/collate.md)

**CBOR to Collate**

* CBOR Types `null`, `true`, `false`, `float32`, `float64`, `integer`,
  `string`, `[]byte` (aka binary), `array`, `object` can be
  collated.
* Indefinite-length encoding for text and binary are not supported.
* LengthPrefix and Stream encoding for array and maps are supported.


**Collate to CBOR**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to CBOR.

Articles
--------

* [Note on sorting](http://prataprc.github.io/sorting-data.html).

How to contribute
-----------------

* Pick an issue, or create an new issue. Provide adequate documentation for
  the issue.
* Assign the issue or get it assigned.
* Work on the code, once finished, raise a pull request.
* Golog is written in [golang](https://golang.org/), hence expected to follow the
  global guidelines for writing go programs.

**Task list**

* [x] Binary collation: transparently handle int64, uint64 and float64.
* [ ] Support for json.Number
* [ ] UTF-8 collation of strings.
* [ ] JSON-pointer.
  - [ ] JSON pointer for looking up within CBOR map.
  - [ ] JSON pointer for looking up within value-map.

Notes
-----

* Don't change the tag number.
* All supplied APIs will panic in case of error, applications can
  recover from panic, dump a stack trace along with input passed on to
  the API, and subsequently handle all such panics as a single valued
  error.
* For now, maximum integer range shall be within int64.
* `Config` instances, and its APIs, are neither re-entrant nor thread safe.

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
