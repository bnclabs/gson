* document `strict` configuration.
* can we use `strict` for using golang's encodeString ?
* improve code documentation and README.
* test cases for cbor_json.go:tag2json().

JSON:

* support for encoding/json.Number

CBOR:

* make cbor date-time parsing format configurable for tagDateTime.
* integrate CBOR test vector with gson.
* document the different between length prefix and size prefix.
* in transforming cbor to json, encodeString() optimized version of golang's
  encoding/json library is used (contributed by Sarath). keep it in sync with
  upstream (golang's stdlib).
* support for cbor tags: tagBase64URL, tagBase64, tagBase16

Collate:

* create a new directory testdata/collate_len/ that contains the sorted list of json
  items without using `lenprefix`
* utf8 collation for strings.

JsonPointer for Value:

* implement json pointer op PREPEND.

JsonPointer for CBOR:

* document lookup APIs for CBOR.

planned features:

* schema on top of CBOR.
* json patch specification RFC-6902.

rules for protocol upgrade:

* don't change the tag number.
* don't have mandatory fields.
