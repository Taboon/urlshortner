package entity

import "errors"

var ErrUnknownID = errors.New("unknown ID")
var ErrEmptyFlag = errors.New("пустое значение флага")