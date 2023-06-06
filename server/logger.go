package server

import (
	"fmt"
	"k8s.io/klog/v2"
)

type Logger interface {
	Info(msg string)
	Infof(msg string, vals ...interface{})
	Errorf(msg string, vals ...interface{})
	Fatalf(msg string, vals ...interface{})
}

type KLoggerWrapper struct {
	Prefix string
}

func NewKLoggerWrapper() Logger {
	return &KLoggerWrapper{
		Prefix: fmt.Sprintf("ingress-j8a [%v] ", Version),
	}
}

func (k *KLoggerWrapper) Info(msg string) {
	klog.Info(k.Prefix + msg)
}

func (k *KLoggerWrapper) Infof(msg string, vals ...interface{}) {
	klog.Infof(k.Prefix+msg, vals...)
}

func (k *KLoggerWrapper) Errorf(msg string, vals ...interface{}) {
	klog.Errorf(k.Prefix+msg, vals...)
}

func (k *KLoggerWrapper) Fatalf(msg string, vals ...interface{}) {
	klog.Fatalf(k.Prefix+msg, vals...)
}
