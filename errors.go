package cache

import "github.com/Nigel2392/errors"

const (
	CodeItemNotFound errors.GoCode = "ItemNotFound"
	CodeInvalidType  errors.GoCode = "InvalidType"
	CodeNotSupported errors.GoCode = "UnsupportedOperation"
)

var (
	ErrItemNotFound = errors.New(CodeItemNotFound, "item not found")
	ErrInvalidType  = errors.New(CodeInvalidType, "invalid type received")
	ErrNotSupported = errors.New(CodeNotSupported, "unsupported operation")
)
