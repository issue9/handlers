// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"
)

// 同时实现了http.ResponseWriter和http.Hijacker接口。
type compressWriter struct {
	gzw io.Writer
	rw  http.ResponseWriter
	hj  http.Hijacker
}

func (cw *compressWriter) Write(bs []byte) (int, error) {
	h := cw.rw.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(bs))
	}

	return cw.gzw.Write(bs)
}

func (cw *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return cw.hj.Hijack()
}

func (cw *compressWriter) Header() http.Header {
	return cw.rw.Header()
}

func (cw *compressWriter) WriteHeader(code int) {
	cw.rw.WriteHeader(code)
}

type compress struct {
	h http.Handler
}

// 支持gzip或是deflate功能的handler。
// 根据客户端请求内容自动匹配相应的压缩算法，优先匹配gzip。
//
// 经过压缩的内容，可能需要重新指定Content-Type，系统检测的类型未必正确。
func Compress(h http.Handler) *compress {
	return &compress{h: h}
}

func CompressFunc(f func(http.ResponseWriter, *http.Request)) *compress {
	return &compress{h: http.HandlerFunc(f)}
}

func (c *compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		hj = nil
	}

	var gzw io.WriteCloser
	var encoding string
	encodings := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	for _, encoding = range encodings {
		encoding = strings.ToLower(strings.TrimSpace(encoding))

		if encoding == "gzip" {
			gzw = gzip.NewWriter(w)
			break
		}

		if encoding == "deflate" {
			var err error
			gzw, err = flate.NewWriter(w, flate.DefaultCompression)
			if err != nil { // 若出错，不压缩，直接返回
				c.h.ServeHTTP(w, r)
				return
			}
			break
		}
	} // end for
	if gzw == nil { // 不支持的压缩格式
		return
	}

	w.Header().Set("Content-Encoding", encoding)
	w.Header().Add("Vary", "Accept-Encoding")
	cw := &compressWriter{
		gzw: gzw,
		rw:  w,
		hj:  hj,
	}

	// 只要gzw!=nil的，必须会执行到此处。
	defer gzw.Close()

	// 此处可能panic，所以得保证在panic之前，gzw变量已经Close
	c.h.ServeHTTP(cw, r)
}
