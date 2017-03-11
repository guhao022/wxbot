package log

import (
	"github.com/num5/loger"
)

var log *loger.Log

// DEBUG
func Debug(v ...interface{}) {
	log.Debug(v)
}

func Debugf(format string, v ...interface{}) {
	log.Debugf(format, v)
}

// Trace
func Trac(v ...interface{}) {
	log.Trac(v)
}

func Tracf(format string, v ...interface{}) {
	log.Tracf(format, v)
}

// INFO
func Info(v ...interface{}) {
	log.Info(v)
}

func Infof(format string, v ...interface{}) {
	log.Infof(format, v)
}

//WARNING
func Warn(v ...interface{}) {
	log.Warn(v)
}

func Warnf(format string, v ...interface{}) {
	log.Warnf(format, v)
}

// ERROR
func Error(v ...interface{}) {
	log.Error(v)
}

func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v)
}

// FATAL
func Fatal(v ...interface{}) {
	log.Fatal(v)
}

func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v)
}

func init() {

	// 初始化
	log = loger.NewLog(1000)
	// 设置输出引擎
	log.SetEngine("file", `{"level":4,"spilt":"size", "filename":".webot/logs/test.log", "maxsize":10}`)

	//log.DelEngine("console")

	// 设置是否输出行号
	log.SetFuncCall(true)
	log.SetFuncCallDepth(5)

	// 设置log级别
	//log.SetLevel("Warning")
}
