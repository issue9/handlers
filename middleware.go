// SPDX-License-Identifier: MIT

// Package middleware 包含了一系列 http.Handler 接口的中间件
package middleware

import "net/http"

// Middleware 将一个 http.Handler 封装成另一个 http.Handler
type Middleware func(http.Handler) http.Handler

// Middlewarer 声明了将对象转换成中间件的接口
type Middlewarer interface {
	// 将当前对象的功能应用于 http.Handler
	Middleware(http.Handler) http.Handler

	// 将当前对象的功能应用于 http.HandlerFunc
	MiddlewareFunc(func(http.ResponseWriter, *http.Request)) http.Handler
}

// Handler 将所有的中间件应用于 h
//
// 后添加的 middleware 会先执行。
func Handler(h http.Handler, middleware ...Middleware) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

// HandlerFunc 将所有的中间件应用于 h
//
// 后添加的 middleware 会先执行。
func HandlerFunc(h func(w http.ResponseWriter, r *http.Request), middleware ...Middleware) http.Handler {
	return Handler(http.HandlerFunc(h), middleware...)
}
