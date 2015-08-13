package cbor

import "errors"

// ErrorDecodeInfoReserved cannot decoding reserved info values 28,29,30.
var ErrorDecodeInfoReserved = errors.New("cbor.decodeInfoReserved")

// ErrorDecodeSimpleType cannot decode invalid simple type.
var ErrorDecodeSimpleType = errors.New("cbor.decodeSimpleType")

// ErrorDecodeIndefinite cannot decode indefinite item.
var ErrorDecodeIndefinite = errors.New("cbor.decodeIndefinite")

// ErrorExpectedCborPointer expect a cbor-pointer
var ErrorExpectedCborPointer = errors.New("cbor.expectedCborPointer")

// ErrorExpectedJsonPointer expect a json-pointer
var ErrorExpectedJsonPointer = errors.New("cbor.expectedJsonPointer")

// ErrorEmptyPointer json-pointer is "", not supported
var ErrorEmptyPointer = errors.New("cbor.errorEmptyPointer")
