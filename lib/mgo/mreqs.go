package mgo

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// HttpServe run a http server
// addr format 0.0.0.0:8080
// performance test tool and setting
// **socket: too many open files**
//   ulimit -a | grep open      // shows 1024
//   ulimit -n 65535            // set a large number
//   lsof -n|awk '{print $2}'|sort|uniq -c|sort -nr|head // list num pid
// **connect: cannot asign requested address
//   cat /proc/net/sockstat     // shows socket usage status
// **apr_socket_recv: Connection reset by peer**
//   server has too many connections possible syn flooding
//   set server /etc/sysctl.conf: net.ipv4.tcp_syncookies = 0
// **profile using pprof and graphviz**
//   go tool pprof ./binary URL/debug/pprof/profile // CPU-profile, MEM-heap
//   usful cmd: top10/web [func]/list [func]
// **useful mux/handler**
//   NewServeMux() // create mux replace DefaultServeMux
//   FileServer, NotFoundHandler, RedirectHandler
func HttpServe(addr string, route func(w http.ResponseWriter, r *http.Request)) error {
	http.HandleFunc("/", route)
	return http.ListenAndServe(addr, nil)
}

// HttpPost post request to url
func HttpPost(url string, reqs, resp interface{}, timeout int) error {
	body, err := json.Marshal(reqs)
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	raw, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer raw.Body.Close()

	body, err = ioutil.ReadAll(raw.Body)
	if err != nil {
		return err
	}

	if len(body) != 0 && resp != nil {
		err = json.Unmarshal(body, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// HttpDelete run delete method to url
func HttpDelete(url string, resp interface{}, timeout int) error {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	raw, err := client.Do(request)
	if err != nil {
		return err
	}
	defer raw.Body.Close()

	body, err := ioutil.ReadAll(raw.Body)
	if err != nil {
		return err
	}

	if len(body) != 0 && resp != nil {
		err = json.Unmarshal(body, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// RateLimiter can be used to limit request rate
type RateLimiter struct {
	ticker *time.Ticker  // const duration ticker
	yield  int64         // const yield balls each duration
	limit  int64         // const bucket size
	balls  int64         // atomic current available balls
	lock   sync.Mutex    // lock allow function include check+set two step
	ch     chan struct{} // channel used to wait balls
}

// NewRateLimiter create a rate-limiter
func NewRateLimiter(dur time.Duration, yield, limit int64) *RateLimiter {
	r := &RateLimiter{
		ticker: time.NewTicker(dur),
		yield:  yield,
		limit:  limit,
		balls:  0,
		ch:     make(chan struct{}),
		lock:   sync.Mutex{},
	}
	go r.run()
	return r
}

func (self *RateLimiter) run() {
	for range self.ticker.C {
		if atomic.LoadInt64(&self.balls)+self.yield < self.limit {
			atomic.AddInt64(&self.balls, self.yield)
		} else {
			atomic.StoreInt64(&self.balls, self.limit)
		}

		for atomic.LoadInt64(&self.balls) > 0 {
			select {
			case <-self.ch:
				atomic.AddInt64(&self.balls, -1)
			default:
				break
			}
		}
	}
}

// Allow return if has balls
func (self *RateLimiter) Allow() bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	if atomic.LoadInt64(&self.balls) < 1 {
		return false
	}
	atomic.AddInt64(&self.balls, -1)
	return true
}

// Wait will wait until has balls
func (self *RateLimiter) Wait() {
	if self.Allow() {
		return
	}

	var s struct{}
	self.ch <- s
}

// InitLimiter init global limiter
func InitLimiter(dur time.Duration, yield, limit int64) {
	once.Do(func() {
		Glimiter = NewRateLimiter(dur, yield, limit)
	})
}

// LimiterAllow check global limiter
func LimiterAllow() bool {
	if Glimiter == nil {
		InitLimiter(10*time.Millisecond, 10, 100)
	}
	return Glimiter.Allow()
}

// LimitWait wait on global limiter
func LimiterWait() {
	if Glimiter == nil {
		InitLimiter(10*time.Millisecond, 10, 100)
	}
	Glimiter.Wait()
}
