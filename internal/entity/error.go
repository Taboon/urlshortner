package entity

import "errors"

var ErrUnknownID = errors.New("unknown ID")
var ErrEmptyFlag = errors.New("empty flag")
var ErrEmptyAnswerDB = errors.New("empty answer DB")
var ErrURLInvalid = errors.New("has no prefix http:// or https://")
var ErrIsNoURL = errors.New("is no url")
var ErrURLExist = errors.New("url already exist")
