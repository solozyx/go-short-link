package main

import (
	"log"
	"net/http"
	"time"
)

type Middleware struct {
}

// 记录请求所消耗的时间
func (m Middleware) LoggingHandler(next http.Handler) http.Handler {
	// 定义 fn 匿名函数,实现闭包功能
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		// next 下1个Handler, ServeHTTP 是对1个http请求的响应
		// 在闭包内对原有 ServeHTTP 函数做装饰,记录前后时间 t1 t2
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v", r.Method, r.URL.String(), t2.Sub(t1))
	}
	// http.HandlerFunc 是适配器,把自定义函数转为 http handler
	return http.HandlerFunc(fn)
}

// 把程序从panic中恢复 避免程序crash
func (m Middleware) RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// 当父函数退出之前执行 defer 的函数
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recover from panic : %+v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
