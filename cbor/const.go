package cbor

import "errors"

// ErrorUnknownType to encode
var ErrorUnknownType = errors.New("cbor.unknownType")

// ErrorExpectedInteger to encode
var ErrorExpectedInteger = errors.New("cbor.expectedInteger")

// ErrorEmptyText to scan
var ErrorEmptyText = errors.New("cbor.emptyText")

// ErrorExpectedNil expected a `nil` token while scanning.
var ErrorExpectedNil = errors.New("cbor.exptectedNil")

// ErrorExpectedTrue expected a `true` token while scanning.
var ErrorExpectedTrue = errors.New("cbor.exptectedTrue")

// ErrorExpectedFalse expected a `false` token while scanning.
var ErrorExpectedFalse = errors.New("cbor.exptectedFalse")

// ErrorExpectedClosearray expected a `]` token while scanning.
var ErrorExpectedClosearray = errors.New("cbor.exptectedCloseArray")

// ErrorExpectedKey expected a `key-string` token while scanning.
var ErrorExpectedKey = errors.New("cbor.exptectedKey")

// ErrorExpectedColon expected a `:` token while scanning.
var ErrorExpectedColon = errors.New("cbor.exptectedColon")

// ErrorExpectedCloseobject expected a `}` token while scanning.
var ErrorExpectedCloseobject = errors.New("cbor.exptectedCloseobject")

// ErrorExpectedToken expected a valid json token while scanning.
var ErrorExpectedToken = errors.New("cbor.exptectedToken")

// ErrorExpectedString expected a `string` token while scanning.
var ErrorExpectedString = errors.New("cbor.exptectedString")

// ErrorInfoReserved for info values 28,29,30.
var ErrorInfoReserved = errors.New("cbor.infoReserved")

// ErrorInfoIndefinite for info indefinite.
var ErrorInfoIndefinite = errors.New("cbor.infoIndefinite")

// ErrorUnassigned for info indefinite.
var ErrorUnassigned = errors.New("cbor.unassigned")

// ErrorByteString byte string decoding not supported for cbor->json.
var ErrorByteString = errors.New("cbor.byteString")

// ErrorExpectedIndefinite for arrays and maps for cbor->json.
var ErrorExpectedIndefinite = errors.New("cbor.expectedIndefinite")

// ErrorTagNotSupported for arrays and maps for cbor->json.
var ErrorTagNotSupported = errors.New("cbor.tagNotSupported")

// ErrorUndefined cannot decode simple-type undefined.
var ErrorUndefined = errors.New("cbor.undefined")

// ErrorSimpleType unsupported simple-type.
var ErrorSimpleType = errors.New("cbor.simpleType")

// ErrorFloat16 simple type not supported.
var ErrorFloat16 = errors.New("cbor.float16")

// ErrorUnexpectedText should be prefixed by tagJsonString.
var ErrorUnexpectedText = errors.New("cbor.unexpectedText")

// ErrorBreakcode simple type not supported with breakcode.
var ErrorBreakcode = errors.New("cbor.breakcode")

// ErrorExpectedCborPointer expect a cbor-pointer
var ErrorExpectedCborPointer = errors.New("cbor.expectedCborPointer")

// ErrorExpectedJsonPointer expect a json-pointer
var ErrorExpectedJsonPointer = errors.New("cbor.expectedJsonPointer")

// ErrorInvalidArrayOffset
var ErrorInvalidArrayOffset = errors.New("cbor.invalidArrayOffset")

// ErrorInvalidPointer
var ErrorInvalidPointer = errors.New("cbor.invalidPointer")

// ErrorNoKey
var ErrorNoKey = errors.New("cbor.noKey")

// ErrorMalformedDocument
var ErrorMalformedDocument = errors.New("cbor.malformedDocument")

// ErrorInvalidDocument
var ErrorInvalidDocument = errors.New("cbor.invalidDocument")
