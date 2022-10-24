package api

import (
	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type logger int

var _ resty.Logger = logger(0)

func (logger) Debugf(message string, args ...interface{}) {
	log.Debugf(message, args...)
}

func (logger) Errorf(message string, args ...interface{}) {
	log.Errorf(message, args...)
}

func (logger) Warnf(message string, args ...interface{}) {
	log.Warnf(message, args...)
}
