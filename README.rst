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
  accessing into json text, and after unquoting segments within
  the pointer shall be binary compared with property keys.
* json-pointers can be used to access gson or cbor representation.

**gson**

* golang object parsed from json, cbor or collate representation.
* json arrays are represeted in golang as ``[]interface{}``.
* json objects, aka properties, are presented in golang as
  ``map[string]interface{}``.
* gson objects do support operations like, Get(), Set(), and
  Delete() on individual fields located by the json-pointer.

**cbor**

* Concise Binary Object Representation, also called CBOR, RFC-7049.
* machine friendly, designed for inter-networking of light weight
  devices, and easy to implement in many languages.
* can be used for more than data exchange, left to user imagination.

**collate**

* a custom encoding based on a paper and improvised to handle
  JSON specification.
* binary representation preserving the sort order.
* transform back to original JSON from binary representation.
* numbers can be encoded in three way as integers, or as
  small-decimals, or as floating point represented in exponent
  form.
* strings are collated as it is received from the input **without
  un-quoting** the JSON-string and **without unicode collation**.
* strings are byte stuffed to escape item Terminator.
* items in a property object are sorted by its property name
  before they are compared with property's value.
* per couchbase-N1QL requirement collation of missing type, a
  unit type, is also supported.

Transforms
----------

**json to gson**

* a custom parser is supplied that must be faster than encoding/json.
* numbers can be interpreted as integer, or float64, or retained as
  string based on the configuration parameter ``NumberKind``.

  a. ``StringNumber``, to retain string as JSON string type aliased
     to ``encoding/json.Number``, a custom type defined by this package.
     Can be used for delayed parsing.
  b. ``IntNumber``, to interpret JSON number as integer size defined
     by the platform.
  c. ``FloatNumber``, to interpret JSON number as 64-bit floating point.

* whitespace can be interpreted, based on configuration parameter
  ``SpaceKind``, as ``AnsiSpace`` that should be faster
  than ``UnicodeSpace``, while the later supports unicode whitespaces
  as well.
* to convert gson back to json text encoding/json package golang's
  stdlib can be used.

**gson to cbor**

* ``nil``, ``true``, ``false`` golang types are encodable into cbor
  format.
* all golang ``number`` types are encodable into cbor format.
* ``[]byte`` is encoded as cbor byte-string.
* ``string`` is encoded as cbor text.
* generic ``array`` is interpreted as golang ``[]interface{}`` and
  encoded as cbor array.

  a. with ``LengthPrefix`` option for ContainerEncoding, arrays and
     maps are encoded with its length.
  b. with ``Stream`` option, arrays and maps are encoded using
     Indefinite and Breakstop encoding.

* generic ``property`` is interpreted as golang ``[][2]interface{}``
  and encoded as cbor array of 2-element item, where the first item
  is key represented as string and second item is any valid json
  value.
* following golang data types are encoded using cbor-tags,

  a. ``time.Time`` encoded with tag-0.
  b. ``gson/cbor.Epoch``, type supplied by cbor package, encoded
     with tag-1.
  c. ``gson/cbor.EpochMicro``, type supplied by cbor package, encoded
     with tag-1.
  d. ``math/big.Int`` positive numbers are encoded with tag-2, and
     negative numbers are encoded with tag-3.
  e. ``gson/cbor.DecimalFraction``, type supplied by cbor package,
     encoded with tag-4.
  f. ``gson/cbor.BigFloat``, type supplied by cbor package, encoded
     with tag-5.
  g. ``gson/cbor.Cbor``, type supplied by cbor package, encoded with
     tag-24.
  h. ``regexp.Regexp`` encoded with tag-35.
  i. ``gson/cbor.CborPrefix``, type supplied by cbor package, encoded
     with tag-55799.

* all other types shall cause a panic.

**cbor to gson**

* reverse of all ``gson to cbor`` encoding, described above, are
  supported.
* cannot decode ``float16`` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks, array and map shall
  be decoded outside this package using
  ``IsIndefinite*()`` and ``IsBreakcode*()`` APIs.

**json to cbor**

* ``nil``, ``true``, ``false`` golang types are encodable into cbor
  format.
* ``number`` types are encoded based on configuration parameter
  ``NumberKind``, which can be one of the following.

  a. for ``JsonNumber`` number is encoded as cbor-text
     (aka cbor-String) and the whole item is tagged as
     ``tagJsonNumber`` (tag-37).
  b. for ``FloatNumber`` number is encoded as cbor-float64.
  c. for ``FloatNumber32`` number is encoded as cbor-float32.
  d. for ``IntNumber`` number is encoded as cbor-int64.
  e. for ``SmartNumber`` number is floating point then it is
     encoded as cbor-float64, else encoded as cbor-int64.
  f. SmartNumber32 number is floating point then it is encoded
     as cbor-float32, else encoded as cbor-float32.

* ``string`` is encoded as cbor-text and the whole item is tagged
  as ``tagJsonString`` (tag-37).
* ``array`` is encoded as indefinite cbor-array which shall be
  finalized using breakstop.
* ``property`` is encoded as indefinite cbor-map which shall be
  finalized using breakstop, property key shall be encoded as
  cbor-text and the whole item is tagged as ``tagJsonString``.

**cbor to json**

* ``nil``, ``true``, ``false`` cbor types are transformed back to
  equivalent json types.
* ``float32`` and ``float64`` are transformed back to 32 bit
  string and 64 bit string respectively in non-exponent format.
* ``integers`` are transformed back to string representation of
  of number and integers exceeding 9223372036854775807 are not
  supported.
* ``tagJsonNumber`` are interpreted as it is into JSON number.
* ``strings`` as utf8-encoded string (aka cbor-text) and JSON
  strings that are tagged using ``tagJsonString`` are interpreted
  as it is.

  a. **note that cbor-text in utf8 format won't be quoted or
     escaped into JSON string**.

* ``arrays`` encoded with length prefix and with indefinite
  encoding are converted to json array.
* ``maps`` encoded with length prefix and with indefinite
  encoding are converted to json property.
* bytes-strings are not supported or transformed to json.
* cbor-text with indefinite encoding are not supported.
* simple type float16 are not supported.

**gson to collate**

* TBD

**json to collate**

* TBD

**cbor to collate**

* TBD

**collate to gson**

* TBD

**collate to json**

* TBD

**collate to cbor**

* TBD

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

Notes
-----

* All supplied APIs will panic in case of error, applications can
  recover from panic, dump a stack trace along with input passed on to
  the API, and subsequently handle all such panics as a single valued
  error.
* maximum integer space shall be int64, uint64 is not supported.

for api documentation and bench marking try,

.. code-block:: bash

    godoc github.com/couchbaselabs/go-collatejson | less
    cd go-collatejson
    go test -test.bench=.

to measure relative difference in sorting 100K elements using encoding/json
library and this library try,

.. code-block:: bash

    go test -test.bench=Sort

examples/* contains reference sort ordering for different json elements.

For known issues refer to `TODO.rst`
