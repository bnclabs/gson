* encodeString() make it default and also provide Golang complaint version.
* integrate CBOR test vector with gson.
* validate collation.
* from json->cbor support LengthPrefix encoding.
* implement json pointer op PREPEND for a gson document.
* create a new directory examples_len/ that contains the sorted list of json
  items without using `lenprefix`
* make cbor date-time parsing format configurable for tagDateTime.

* document lookup APIs for CBOR and GSON.
* support for cbor tags: tagBase64URL, tagBase64, tagBase16
* utf8 collation.

planned features:

* schema on top of CBOR.
* json patch specification RFC-6902.

rules for protocol upgrade:

* don't change the tag number.
* don't have mandatory fields.
