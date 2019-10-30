package main

import (
	"flag"
	"fmt"
	"github.com/kataras/iris/v12"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
)

var (
	url                       string = "127.0.0.1:8080"
	total                     int64
	mode                      string
	count, thread, serverType int
	wg                        sync.WaitGroup
	lock                      sync.Mutex
	totalCh                   chan *int64
)

func delCountSimple() bool {
	if total > 0 {
		total--
		return true
	}
	return false
}

func delCountAtomic() bool {
	if total > 0 {
		if atomic.AddInt64(&total, -1) >= 0 {
			return true
		}
		total = 0
	}
	return false
}

func delCountMutex() bool {
	ret := false
	lock.Lock()
	if total > 0 {
		ret = true
		total--
	}
	lock.Unlock()
	return ret
}

func delCountChannel() bool {
	ret := false
	if *<-totalCh > 0 {
		ret = true
		total--
	}
	totalCh <- &total
	return ret
}

func clientRun() {
	_, err := http.Get(fmt.Sprintf("http://%s/add/%d", url, count))
	if err != nil {
		fmt.Println(err)
		return
	}

	var testTotal, errCount, finished int64

	for i := 0; i < thread; i++ {
		wg.Add(1)
		go func() {
			for {
				resp, err := http.Get(fmt.Sprintf("http://%s/del", url))
				if err != nil {
					atomic.AddInt64(&errCount, 1)
					break
				}

				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)

				if string(body) == "1" {
					atomic.AddInt64(&testTotal, 1)
				} else if string(body) == "0" {
					atomic.AddInt64(&finished, 1)
					break
				} else {
					atomic.AddInt64(&errCount, 1)
					break
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println(testTotal, finished, errCount)
}

func localRun() {
	total = int64(count)

	var testTotal, errCount, finished int64

	for i := 0; i < thread; i++ {
		wg.Add(1)
		go func() {
			for {
				success := false
				switch serverType {
				case 0:
					success = delCountSimple()
				case 1:
					success = delCountAtomic()
				case 2:
					success = delCountMutex()
				case 3:
					success = delCountChannel()
				}

				if success {
					atomic.AddInt64(&testTotal, 1)
				} else {
					atomic.AddInt64(&finished, 1)
					break
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println(testTotal, finished, errCount)
}

func serverRun() {
	app := iris.Default()
	logger := app.Logger()
	logger.SetLevel("error")

	app.Get("/add/{count:int64}", func(ctx iris.Context) {
		total += ctx.Params().GetInt64Default("count", 0)
		ctx.Writef("%d\n", total)
	})

	app.Get("/get", func(ctx iris.Context) {
		ctx.Writef("%d\n", total)
	})

	app.Get("/del", func(ctx iris.Context) {
		success := false
		switch serverType {
		case 0:
			success = delCountSimple()
		case 1:
			success = delCountAtomic()
		case 2:
			success = delCountMutex()
		case 3:
			success = delCountChannel()
		}
		if success {
			ctx.WriteString("1")
		} else {
			ctx.WriteString("0")
		}
	})

	app.Run(iris.Addr(url))
}

func main() {
	flag.StringVar(&mode, "m", "", "`mode`: c=client, s=server, l=local")
	flag.IntVar(&count, "n", 1000, "[client mode] `count` to add")
	flag.IntVar(&thread, "t", 1, "[client mode] `thread`")
	flag.IntVar(&serverType, "x", 1, "[server mode] `type`: 0=simple, 1=atomic, 2=mutex, 3=channel")
	flag.Parse()

	// channel init
	totalCh = make(chan *int64, 1)
	totalCh <- &total

	switch mode {
	case "s":
		serverRun()
	case "c":
		clientRun()
	case "l":
		localRun()
	default:
		flag.PrintDefaults()
	}
}
