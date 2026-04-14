package logger

import (
	"fmt"
	"log"
)

func Info(message string) {
	log.Printf("[LynxPilot][信息]:%s", message)
}

func Error(message string) {
	log.Printf("[LynxPilot][错误]:%s", message)
}

func Infof(format string, args ...any) {
	log.Printf("[LynxPilot][信息]:"+format, args...)
}

func Errorf(format string, args ...any) error {
	log.Printf("[LynxPilot][错误]:"+format, args...)
	return fmt.Errorf(format, args...)
}
