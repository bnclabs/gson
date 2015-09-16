* document lookup APIs for CBOR and GSON.
* add testcase for testdata/collate/
* memory profile using tools/gson program for gson, cbor and collate.
* from json->cbor support LengthPrefix encoding.
* implement json pointer ops SET/DEL/GET/PREPEND for a JSON document.
* implement json pointer ops SET/DEL/GET/PREPEND for a collated document.
* implement json pointer op PREPEND for a gson document.
* utf8 collation.
* The CouchDB collation spec uses Unicode collation, and strangely enough
  the collation order for ASCII characters is not the same as ASCII order. I
  solved this by creating a mapping table that converts the bytes 0-127 into
  their priority in the Unicode collation.
* create a new directory examples_len/ that contains the sorted list of json
  items without using `lenprefix`

planned features:

* schema on top of CBOR.
* json patch specification RFC-6902.

rules for protocol upgrade:

* don't change the tag number.
* don't have mandatory fields.
