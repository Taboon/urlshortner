package entity

import "errors"

var ErrHasNoPrefix = errors.New("has no prefix http:// or https://")
var ErrHasNoDot = errors.New("has no dot in url")

var ErrURLExist = errors.New("url already exist")
var ErrUnknownUser = errors.New("unknown user")

var ErrUnknownID = errors.New("unknown ID")
var ErrEmptyFlag = errors.New("empty flag")

var ErrNoURLToSave = errors.New("no have url to save")
var ErrJSONInvalid = errors.New("invalid json")

var ErrRepositoryNotInitialized = errors.New("repository not initialized")
