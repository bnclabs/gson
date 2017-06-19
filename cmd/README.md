General options
---------------

* ``-inpfile`` file containing one or more json docs based on the context.
* ``-inptxt`` use input text for the operation.
* ``-mprof`` take memory profile for testdata/code.json.gz.
* ``-nk`` can be ``smart``, treat number as int64 or fall back to float64, or,
  ``float``, treat number as float64 (default "float").
* ``-outfile`` write output to file.
* ``-overheads`` compute overheads on cbor and collation encoding.
* ``-quote`` use strconv.Unquote on inptxt/inpfile.
* ``-repeat`` repeat count.
* ``-ws`` can be ``ansi`` whitespace, or, ``unicode`` whitespace, default "ansi".

Convert from JSON
-----------------

* ``-json2cbor`` convert inptxt or content in inpfile to cbor output.
* ``-json2collate`` convert inptxt or content in inpfile to collated output.
* ``-json2value`` convert inptxt or content in inpfile to golang value.

**options for JSON**

* ``-pointers`` list of json-pointers for doc specified by input-file.

Convert from CBOR
-----------------

* ``-cbor2collate`` convert inptxt or content in inpfile to collated output.
* ``-cbor2json`` convert inptxt or content in inpfile to json output.
* ``-cbor2value`` convert inptxt or content in inpfile to golang value.

**options for CBOR**

* ``-ct`` container encoding for cbor, allowed ``stream`` (default), or,
  ``lenprefix``.

Convert from Collate
--------------------

* ``-collate2cbor`` convert inptxt or content in inpfile to cbor output.
* ``-collate2json`` convert inptxt or content in inpfile to json output.


**options for collation**

* ``-arrlenprefix`` set SortbyArrayLen for collation ordering.
* ``-checkdir`` test files for collation order in specified directory. For
  every input file ``<checkdir>/filename``, there should be reference file
  ``<checkdir>/filename.ref``.
* ``-collatesort`` sort inpfile, with one or more JSON terms, using
  collation algorithm.
* ``-n1qlsort`` sort inpfile, with one or more JSON terms, using
  collation algorithm.
* ``-domissing`` consider missing type while collation (default true).
* ``-maplenprefix`` SortbyPropertyLen for collation ordering (default true)

Convert from value
------------------

* ``-value2cbor`` convert inptxt json to value and then to cbor
* ``-value2json`` convert inptxt json to value and then back to json
