package mgo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

// mutex deadlock with dirty
type Logger struct {
	pre  string // log filename prefix, empty pre means log to stdio
	max  int64  // log file max size
	sync bool   // log sync mode

	fio io.ReadWriteCloser // log file handler
	num int64              // log file size, num=-1 indicate fio not open

	buf   []byte     // log buf
	mutex sync.Mutex // lock: buf

	dirty chan bool // log buf is dirty
}

func (self *Logger) close() {
	if self.fio != nil {
		self.fio.Close()
	}
	self.fio = nil
	self.num = -1
}

func (self *Logger) open() error {
	if self.pre == "" {
		return nil
	}

	if self.fio != nil && self.num >= 0 && self.num < self.max {
		return nil
	}

	if self.num > self.max {
		self.close()
	}

	fname := fmt.Sprintf("%s-%s", self.pre, time.Now().Format("2006010215"))
	if PathExist(fname) && FileSize(fname) > self.max - self.num {
		fname = fmt.Sprintf("%s-%s", self.pre, time.Now().Format("2006010215.0405"))
		if PathExist(fname) {
			fname = fmt.Sprintf("%s-%s", self.pre, time.Now().Format("2006010215.0405.000"))
		}
	}
	err := CreateFile(fname)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	self.fio = f
	self.num = 0
	return nil
}
func (self *Logger) openAndLog() {
	if self.pre == "" {
		panic("missing log file")
	}

	err := self.open()
	if err != nil {
		panic(err)
	}

	if self.num < 0 {
		panic("fio not open")
	}

	self.mutex.Lock()
	if len(self.buf) > 0 {
		n, err := self.fio.Write(self.buf)
		if err != nil {
			panic(err)
		}
		self.buf = []byte{}
		self.num += int64(n)
	}
	self.mutex.Unlock()
}

func (self *Logger) flush() error {
	for {
		select {
		case <-self.dirty:
			self.openAndLog()
		}
	}
}
func (self *Logger) Write(info string) error {
	if self.pre == "" {
		fmt.Printf(info)
		return nil
	}

	self.mutex.Lock()
	self.buf = append(self.buf, []byte(info)...)
	self.mutex.Unlock()

	if self.sync {
		self.openAndLog()
		self.close()
	} else {
		self.dirty <- true
	}
	return nil
}

// NewLogger create new logger
func NewLogger(sync bool, pre string, size int64) *Logger {
	logger := &Logger{
		sync:  sync,
		pre:   pre,
		max:   size,
		num:   -1,
		dirty: make(chan bool, SIZE_1K),
	}

	if !sync {
		go logger.flush()
	}

	return logger
}

// InitLogger init global logger
func InitGlogger(sync bool, pre string, size int64) {
	once.Do(func() {
		Glogger = NewLogger(sync, pre, size)
	})
}

func InitLogger(pre string) {
	InitGlogger(true, pre, 100*SIZE_1M)
}

func logwrite(s string) {
	if Glogger == nil {
		InitLogger("")
	}

	Glogger.Write(s + "\n")
}

// support format
// 1. logf(s)
// 2. logf(fmt, v...)
// 3. logf(any, v...)
func logf(level string, f interface{}, args ...interface{}) string {
	fs, ok := f.(string)
	if !ok {
		fs = fmt.Sprintf("%v", f) + strings.Repeat(" [%v]", len(args))
	}

	pc, file, line, _ := runtime.Caller(2)
	key := fmt.Sprintf("[%s] %s %s:%s():%d ", time.Now().Format(TIME_FORMAT),
		level, path.Base(file), runtime.FuncForPC(pc).Name(), line)
	if uuid := GetUuid(); uuid != "" {
		key += fmt.Sprintf("%s (%d) ", uuid, UuidCacheSize())
	}

	s := strings.TrimSpace(fmt.Sprintf(fs, args...))
	logwrite(key + s)
	return s
}

// Debugf log with debug level
func Debugf(f interface{}, args ...interface{}) {
	logf("DEBUG", f, args...)
}

// Infof log with info level
func Infof(f interface{}, args ...interface{}) {
	logf("INFOF", f, args...)
}

// Errorf log with error level
func Errorf(f interface{}, args ...interface{}) error {
	return errors.New(logf("ERROR", f, args...))
}

// Fatalf log with fatal level
func Fatalf(f interface{}, args ...interface{}) {
	panic(logf("FATAL", f, args...))
}
