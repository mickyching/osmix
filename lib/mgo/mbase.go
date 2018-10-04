package mgo

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"
)

// StartProfile call pprof to start profile
// StartProfile when profiling already enabled will panic.
func StartProfile(fcpu string) {
	f, err := os.Create(fcpu)
	if err != nil {
		Fatalf(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		Fatalf(err)
	}
}

// StopProfile stop profiling
func StopProfile() {
	pprof.StopCPUProfile()
}

// MemProfile write mem info to file, can be called at any time.
func MemProfile(fmem string) {
	f, err := os.Create(fmem)
	if err != nil {
		Fatalf(err)
	}
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		Fatalf(err)
	}
	f.Close()
}

// RunCmd exec cmd with args
func RunCmd(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	return strings.TrimSpace(string(out)), err
}

// PathExist check if path exist
func PathExist(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

// CreateDir create a directory
func CreateDir(dname string) error {
	return os.MkdirAll(dname, os.ModePerm)
}

// CreateFile create a file
func CreateFile(fname string) error {
	if PathExist(fname) {
		return nil
	}
	err := CreateDir(path.Dir(fname))
	if err != nil {
		return err
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// ResetFile try to create and write file
func ResetFile(fname, info string) error {
	if err := CreateFile(fname); err != nil {
		return err
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(info)
	return nil
}

// WriteFile try to create file and append info to file
func WriteFile(fname, info string) error {
	if err := CreateFile(fname); err != nil {
		return err
	}
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(info)
	return nil
}

// GoFunc run func with num concurrency routines
func GoFunc(num int, f func()) *sync.WaitGroup {
	wg := sync.WaitGroup{}
	if num < 1 {
		return &wg
	}

	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			f()
		}()
	}
	return &wg
}

// GoId returns current Goroutine ID
func GoId() int64 {
	gid := func(s []byte) int64 {
		s = s[len("goroutine "):]
		s = s[:bytes.IndexByte(s, ' ')]
		gid, _ := strconv.ParseInt(string(s), 10, 64)
		return gid
	}
	var buf [64]byte
	return gid(buf[:runtime.Stack(buf[:], false)])
}

// Uuid returns uuid base on current time
func Uuid() string {
	unix32bits := uint32(time.Now().UTC().Unix())
	buff := make([]byte, 12)
	numRead, err := rand.Read(buff)
	if numRead != len(buff) || err != nil {
		Fatalf(err)
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x-%x", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:])
}

// UuidCacheSize returns uuid-cache size
func UuidCacheSize() int {
	UuidMutex.RLock()
	num := len(UuidCache)
	UuidMutex.RUnlock()
	return num
}

// SetUuid set uuid
func SetUuid(uuid string) {
	UuidMutex.Lock()
	UuidCache[GoId()] = uuid
	UuidMutex.Unlock()
}

// GetUuid returns current goroutine's uuid
func GetUuid() string {
	UuidMutex.RLock()
	uuid := UuidCache[GoId()]
	UuidMutex.RUnlock()
	return uuid
}

// DelUuid delete current goroutine's uuid
func DelUuid() {
	UuidMutex.Lock()
	defer UuidMutex.Unlock()
	delete(UuidCache, GoId())
}

// Lio is line-based bufio
type Lio struct {
	r *bufio.Scanner
	w *bufio.Writer
}

// NewLio return line-io
func NewLio(f interface{}) *Lio {
	l := new(Lio)
	bufsize := 128 * 1024
	if r, ok := f.(io.Reader); ok {
		buf := make([]byte, bufsize)
		l.r = bufio.NewScanner(r)
		l.r.Buffer(buf, bufsize)
		l.r.Split(bufio.ScanLines)
	}
	if w, ok := f.(io.Writer); ok {
		l.w = bufio.NewWriterSize(w, bufsize)
	}
	return l
}

// Read read line to lio
func (self *Lio) Read() bool {
	return self.r.Scan()
}

// Line get line from lio
func (self *Lio) Line() string {
	return self.r.Text()
}

// Write write line to lio
func (self *Lio) Write(line string) {
	self.w.WriteString(line + "\n")
}

// Write write line to lio
func (self *Lio) Writef(line string, args ...interface{}) {
	self.w.WriteString(fmt.Sprintf(line, args...) + "\n")
}

// Flush flush lio to file
func (self *Lio) Flush() {
	self.w.Flush()
}
