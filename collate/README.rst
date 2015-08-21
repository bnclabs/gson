Collatejson library, written in golang, provides encoding and decoding function
to transform JSON text into binary representation without loosing information.
That is,

* binary representation should preserve the sort order such that, sorting
  binary encoded json documents much match sorting by functions that parse
  and compare JSON documents.
* it must be possible to get back the original document, in semantically
  correct form, from its binary representation.

The following is the list of changes from github.com/prataprc/collatejson,

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
