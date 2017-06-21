Object formats and notations
============================

[![Build Status](https://travis-ci.org/prataprc/gson.png)](https://travis-ci.org/prataprc/gson)
[![Coverage Status](https://coveralls.io/repos/prataprc/gson/badge.png?branch=master&service=github)](https://coveralls.io/github/prataprc/gson?branch=master)
[![GoDoc](https://godoc.org/github.com/prataprc/gson?status.png)](https://godoc.org/github.com/prataprc/gson)

* [What is what](#what-is-what)
* [Transforms](#transforms)
* [Understanding collation](docs/collate.md)
* [Getting started](docs/gettingstarted.md)
* [Play with command line](cmd/README.md)
* [Articles related to gson](#articles)
* [Notes](#notes)

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
* Gson objects support operations like, Get(), Set(), and
  Delete() on individual fields located by the JSON-POINTER.
* Following golang-types can be transformed to JSON, CBOR, or,
  Binary-collation: ``nil``, ``bool``,
  ``byte, int8, int16, uint16, int32, uint32, int, uint, int64, uint64``,
  ``float32, float64``,
  ``string``, ``[]interface{}``, ``map[string]interface{}``,
  ``[][2]interface{}``.
* First item in ``[][2]interface{}`` is treated as key (string) and second
  item is treated as value, hence equivalent to ``map[string]interface{}``

**CBOR**

* Concise Binary Object Representation, also called [CBOR](http://cbor.io/),
  [RFC-7049](https://tools.ietf.org/html/rfc7049).
* Machine friendly, designed for IoT, inter-networking of light weight
  devices, and easy to implement in many languages.
* Can be used for more than data exchange, left to user
  imagination :) ...

**Binary-Collation**

* A custom encoding based on a [paper](docs/pdf) and improvised to
  handle JSON specification.
* Binary representation preserving the sort order.
* Transform back to original JSON from binary representation.
* Numbers can be encoded in three ways - as integers, or as
  small-decimals, or as floating-point represented in exponent
  form.

**JSON-Pointer**

* URL like field locator within a JSON object,
  [RFC-6901](https://tools.ietf.org/html/rfc6901).
* Make sense only for JSON arrays and objects, but to any level
  of nesting.
* JSON-pointers shall be unquoted before they are used for
  accessing into JSON text (or an equivalent representation),
  after unquoting segments within the pointer, each segment shall
  be binary compared with property keys.
* JSON-pointers can be used to access Gson or CBOR representation.
* Documents encoded in CBOR format using LengthPrefix are not
  supported by lookup APIs.

Transforms
----------

![transforms](docs/transforms.png)

**JSON to value**

* Gson uses custom parser that must be faster than encoding/JSON.
* Numbers can be interpreted as integer, or float64,
  -  `IntNumber` to interpret JSON number as integer whose size is
     defined by the platform.
  -  `FloatNumber` to interpret JSON number as 64-bit floating point.

* Whitespace can be interpreted, based on configuration parameter
  `SpaceKind`. SpaceKind can be one of the following `AnsiSpace` or
  `UnicodeSpace`.
  - `AnsiSpace` that should be faster
  - `UnicodeSpace` supports unicode white-spaces as well.

**value to JSON**


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
* Types `number` are encoded based on configuration parameter
  `NumberKind`, which can be one of the following.
  * Type `FloatNumber` number is encoded as CBOR-float64.
  * Type `FloatNumber32` number is encoded as CBOR-float32.
  * Type `IntNumber` number is encoded as CBOR-int64.
  * Type `SmartNumber` if number is floating point then it is encoded
    as CBOR-float64, else encoded as CBOR-int64.
  * Type `SmartNumber32` if number is floating point then it is encoded
    as CBOR-float32, else encoded as CBOR-float32.
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
* Type `strings` is encoded into JSON-string using `encoding/json`
  package.
* Type `array` either with length prefix or with indefinite encoding
  are converted back to JSON array.
* Type `map` either with length prefix or with indefinite encoding
  are converted back to JSON property.
* Type bytes-strings are not supported or transformed to JSON.
* Type CBOR-text with indefinite encoding are not supported.
* Type Simple type float16 are not supported.


**value to collate**

* Types `nil`, `true`, `false`, `float64`, `int64`, `int`, `Missing`,
  `string`, `[]byte`, `[]interface{}`, `map[string]interface{}`
  are supported for collation.
* If configured as `FloatNumber`, `FloatNumber32` number will
  be collated as floating point.
* If configured as `IntNumber` number will be collated as integer.
* If configured as `Decimal` number will be collated as
  small-decimal ( -1 >= num <= 1 ).
* If string value is MissingLiteral, it shall be collated as
  missing.

**collate to value**

* Types `Missing`, `null`, `true`, `false`, `floating-point`,
  `small-decimal`, `integer`, `string`, `[]byte` (aka binary),
  `array`, `object` from its collated from can be converted back
  to value.

**JSON to collate**

* Types `null`, `true`, `false`, `number`, `string`, `array`, `object`
  are supported for collation.
* Type `number` is parsed as float64 and collated based on configuration:
  * If configured as `FloatNumber`, `FloatNumber32` number will be
    collated as floating point.
  * If configured as `IntNumber` number will be collated as integer.
  * If configured as `Decimal` number will be collated as
    small-decimal ( -1 >= num <= 1 ).
* If string value is MissingLiteral, it shall be collated as
  missing.
* All other `string` value will be encoded into UTF-8 format before
  collating it.

**collate to JSON**

* `null`, `true`, `false`, `number`, `string`, `array`, `object`
  types are converted back to JSON.

**CBOR to collate**

* CBOR Types `null`, `true`, `false`, `float32`, `float64`, `integer`,
  `string`, `[]byte` (aka binary), `array`, `object` can be
  collated.
* Indefinite-length encoding for text and binary are not supported.
* LengthPrefix and Stream encoding for array and maps are supported.


**collate to CBOR**

* `Missing`, `null`, `true`, `false`, `floating-point`, `small-decimal`,
  `integer`, `string`, `[]byte` (aka binary), `array`, `object` types
  from its collated from can be converted back to CBOR.

Notes
-----

* Don't change the tag number.
* Don't have mandatory fields.
* All supplied APIs will panic in case of error, applications can
  recover from panic, dump a stack trace along with input passed on to
  the API, and subsequently handle all such panics as a single valued
  error.
* Maximum integer space shall be in int64.
* `Config` instances, and its APIs, are neither re-entrant not thread safe.
* Encoding/json.Number is not supported yet.

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

Task list
=========

* [ ] Binary collation: transparently handle int64, uint64 and float64.
* [ ] Support for json.Number
* [ ] UTF-8 collation of strings.
* [ ] JSON-pointer.
  - [ ] JSON pointer for looking up within CBOR map.
  - [ ] JSON pointer for looking up within value-map.

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
