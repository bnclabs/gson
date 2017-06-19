General options
---------------

* ``-inpfile`` process file containing json doc(s) based on other options.
* ``-inptxt`` process input text based on their options.
* ``-mprof`` take memory profile for testdata/code.json.gz.
* ``-outfile`` write output to file.
* ``-overheads`` compute overheads on cbor and collation encoding.
* ``-quote`` use strconv.Unquote on inptxt/inpfile.
* ``-repeat`` repeat count.
* ``-nk`` can be ``smart``, treat number as int64 or fall back to float64, or,
  ``float``, treat number only as float64 (default "float").
* ``-ws`` can be ``ansi`` whitespace, or, ``unicode`` whitespace, default "ansi".

To include [n1ql](https://www.couchbase.com/products/n1ql), compile it with
``-tags n1ql``.

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
* ``-maplenprefix`` SortbyPropertyLen for collation ordering (default true)
* ``-domissing`` consider missing type while collation (default true).
* ``-collatesort`` sort inpfile, with one or more JSON terms, using
  collation algorithm.
* ``-n1qlsort`` sort inpfile, with one or more JSON terms, using
  collation algorithm.
* ``-checkdir`` test files for collation order in specified directory. For
  every input file ``<checkdir>/filename``, there should be reference file
  ``<checkdir>/filename.ref``.

Convert from value
------------------

* ``-value2cbor`` convert inptxt json to value and then to cbor
* ``-value2json`` convert inptxt json to value and then back to json
