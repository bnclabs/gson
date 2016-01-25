* json.Number can be used ? [+] add an explanation in README
* make this constant:
     strlen, numkeys, itemlen, ptrlen := 1024*1024, 1024, 1024*1024, 1024
* improve code documentation and README.
* by using ``[][2]interface{}`` instead of map[string]interface{} we can
  optimize on heap allocation.

        collt cbor json value
value           ✓    ✓    X
json                 X    ✓
cbor            X         ✓
collt     X

CBOR:

* integrate CBOR test vector with gson.
* from json to cbor support LengthPrefix encoding.
* make cbor date-time parsing format configurable for tagDateTime.
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
