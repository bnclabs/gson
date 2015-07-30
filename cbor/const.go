package cbor

import "errors"

// ErrorDecodeInfoReserved cannot decoding reserved info values 28,29,30.
var ErrorDecodeInfoReserved = errors.New("cbor.decodeInfoReserved")

// ErrorDecodeSimpleType cannot decode invalid simple type.
var ErrorDecodeSimpleType = errors.New("cbor.decodeSimpleType")

// ErrorDecodeFloat16 cannot decode float16.
var ErrorDecodeFloat16 = errors.New("cbor.decodeFloat16")

// ErrorDecodeExceedInt64 cannot decode float16.
var ErrorDecodeExceedInt64 = errors.New("cbor.decodeExceedInt64")

// ErrorJsonEmpty to scan
var ErrorJsonEmpty = errors.New("cbor.jsonEmpty")

// ErrorDecodeIndefinite cannot decode indefinite item.
var ErrorDecodeIndefinite = errors.New("cbor.decodeIndefinite")

// ErrorExpectedJsonInteger expected a `number` while scanning.
var ErrorExpectedJsonInteger = errors.New("cbor.expectedJsonInteger")

// ErrorExpectedJsonNil expected a `nil` token while scanning.
var ErrorExpectedJsonNil = errors.New("cbor.exptectedJsonNil")

// ErrorExpectedJsonTrue expected a `true` token while scanning.
var ErrorExpectedJsonTrue = errors.New("cbor.exptectedJsonTrue")

// ErrorExpectedJsonFalse expected a `false` token while scanning.
var ErrorExpectedJsonFalse = errors.New("cbor.exptectedJsonFalse")

// ErrorExpectedJsonClosearray expected a `]` token while scanning.
var ErrorExpectedJsonClosearray = errors.New("cbor.exptectedJsonCloseArray")

// ErrorExpectedJsonKey expected a `key-string` token while scanning.
var ErrorExpectedJsonKey = errors.New("cbor.exptectedJsonKey")

// ErrorExpectedJsonColon expected a `:` token while scanning.
var ErrorExpectedJsonColon = errors.New("cbor.exptectedJsonColon")

// ErrorExpectedJsonCloseobject expected a `}` token while scanning.
var ErrorExpectedJsonCloseobject = errors.New("cbor.exptectedJsonCloseobject")

// ErrorExpectedJsonToken expected a valid json token while scanning.
var ErrorExpectedJsonToken = errors.New("cbor.exptectedJsonToken")

// ErrorExpectedJsonString expected a `string` token while scanning.
var ErrorExpectedJsonString = errors.New("cbor.exptectedJsonString")

// ErrorExpectedJsonIndefinite for arrays and maps while scanning.
var ErrorExpectedJsonIndefinite = errors.New("cbor.expectedJsonIndefinite")

// ErrorByteString byte string decoding not supported for cbor->json.
var ErrorByteString = errors.New("cbor.byteString")

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

// ErrorUnknownType to encode
var ErrorUnknownType = errors.New("cbor.unknownType")
