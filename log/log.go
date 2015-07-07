package log

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/jxwr/cc/streams"
)

// Levels:
// - INFO
// - WARNING
// - ERROR
// - FATAL
// - EVENT

var LogRingBuffer []string

func init() {
	LogRingBuffer = make([]string, 40000)
}

func WriteFileHandler(i interface{}) bool {
	data := i.(*streams.LogStreamData)
	switch data.Level {
	case "VERBOSE":
		glog.V(5).Infof("[%s] %s", data.Target, data.Message)
	case "INFO":
		glog.Infof("[%s] %s", data.Target, data.Message)
	case "WARNING":
		glog.Warningf("[%s] %s", data.Target, data.Message)
	case "ERROR":
		glog.Errorf("[%s] %s", data.Target, data.Message)
	case "FATAL":
		glog.Fatalf("[%s] %s", data.Target, data.Message)
	case "EVENT":
		glog.Infof("[E] [%s] %s", data.Target, data.Message)
	}
	return true
}

func WriteRingBufferHandler(i interface{}) bool {
	msg := i.(*streams.LogStreamData)
	line := fmt.Sprintf("%s %s: [%s] - %s\n", msg.Level,
		msg.Time.Format("2006/01/02 15:04:05"), msg.Target, msg.Message)

	if len(LogRingBuffer) >= cap(LogRingBuffer)-1 {
		LogRingBuffer = LogRingBuffer[1:]
	}
	LogRingBuffer = append(LogRingBuffer, line)

	return true
}

func Verbose(target string, args ...interface{}) {
	level := "VERBOSE"
	message := fmt.Sprint(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Verboseln(target string, args ...interface{}) {
	level := "VERBOSE"
	message := fmt.Sprintln(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Verbosef(target string, format string, args ...interface{}) {
	level := "VERBOSE"
	message := fmt.Sprintf(format, args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Info(target string, args ...interface{}) {
	level := "INFO"
	message := fmt.Sprint(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Infoln(target string, args ...interface{}) {
	level := "INFO"
	message := fmt.Sprintln(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Infof(target string, format string, args ...interface{}) {
	level := "INFO"
	message := fmt.Sprintf(format, args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Warning(target string, args ...interface{}) {
	level := "WARNING"
	message := fmt.Sprint(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Warningln(target string, args ...interface{}) {
	level := "WARNING"
	message := fmt.Sprintln(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Warningf(target string, format string, args ...interface{}) {
	level := "WARNING"
	message := fmt.Sprintf(format, args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Error(target string, args ...interface{}) {
	level := "ERROR"
	message := fmt.Sprint(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Errorln(target string, args ...interface{}) {
	level := "ERROR"
	message := fmt.Sprintln(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Errorf(target string, format string, args ...interface{}) {
	level := "ERROR"
	message := fmt.Sprintf(format, args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Fatal(target string, args ...interface{}) {
	level := "FATAL"
	message := fmt.Sprint(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Fatalln(target string, args ...interface{}) {
	level := "FATAL"
	message := fmt.Sprintln(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Fatalf(target string, format string, args ...interface{}) {
	level := "FATAL"
	message := fmt.Sprintf(format, args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Event(target string, args ...interface{}) {
	level := "EVENT"
	message := fmt.Sprint(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Eventln(target string, args ...interface{}) {
	level := "EVENT"
	message := fmt.Sprintln(args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}

func Eventf(target string, format string, args ...interface{}) {
	level := "EVENT"
	message := fmt.Sprintf(format, args...)
	data := &streams.LogStreamData{
		level, time.Now(), target, message,
	}
	streams.LogStream.Pub(data)
}
