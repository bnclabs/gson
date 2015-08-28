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
* json-pointers will be unquoted before they are used for
  accessing into json text, and after unquoting segments within
  the pointer will be binary compared with property keys.
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

Transforms
----------

**json to gson**

* a custom parser is supplied that must be faster than encoding/json.
* numbers can be interpreted as integer, or float64, or retained as
  string based on the configuration parameter ``NumberKind``.

  a. ``StringNumber``, to retain string as JSON string type aliased
     to ``Number``, a custom type defined by this package. Can be used
     for delayed parsing.
  b. ``IntNumber``, to interpret JSON number as integer size defined
     by the platform.
  c. ``FloatNumber``, to interpret JSON number as 64-bit floating point.

* whitespace can be interpreted, based on configuration parameter
  ``SpaceKind``, as ``AnsiSpace`` that will be faster
  than ``UnicodeSpace``, while the later supports unicode whitespaces
  as well.
* to convert gson back to json text encoding/json package golang's
  stdlib can be used.

**gson to cbor**

* `nil`, `true`, `false` golang types are encodable into cbor format.
* all golang `number` types are encodable into cbor format.
* `[]byte` is encoded as cbor byte-string.
* `string` is encoded as cbor text.
* generic `array` is interpreted as golang ``[]interface{}`` and
  encoded as cbor array.

  a. with ``LengthPrefix`` option for ContainerEncoding, arrays and
     maps are encoded with its length.
  b. with ``Stream`` option, arrays and maps are encoded using
     Indefinite and Breakstop encoding.

* generic `property` is interpreted as golang ``[][2]interface{}`` and
  encoded as cbor array of 2-element item, where the first item is
  key represented as string and second item is any valid json value.
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

* all other types will cause a panic.

**cbor to gson**

* reverse of all `gson to cbor` encoding, described above, are
  supported.
* cannot decode `float16` type and int64 > 9223372036854775807.
* indefinite byte-string chunks, text chunks, array and map shall
  be decoded outside this package using
  ``IsIndefinite*()`` and ``IsBreakcode*()`` APIs.

**json to cbor**

**cbor to json**

Notes
-----

* All supplied APIs will panic in case of error, applications can
  recover from panic, dump a stack trace along with input passed on to
  the API, and subsequently handle all such panics as a single valued
  error.
* items in a property object are sorted by its property name before they
  are compared with other property object.
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
