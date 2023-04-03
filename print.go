package main

import (
	"io"
	"reflect"
	"unsafe"
)

var (
	defaultLogWriter    io.Writer = &serialDevice
	defaultScreenWriter io.Writer = text_mode_writer{}
	defaultErrorWriter  io.Writer = text_mode_error_writer{}
)

// Copied and adapted code from go runtime
func printsp(w io.Writer) {
	printstring(w, " ")
}

func printnl(w io.Writer) {
	printstring(w, "\n")
}

func printbool(w io.Writer, v bool) {
	if v {
		printstring(w, "true")
	} else {
		printstring(w, "false")
	}
}

func printfloat(w io.Writer, v float64) {
	switch {
	case v != v:
		printstring(w, "NaN")
		return
	case v+v == v && v > 0:
		printstring(w, "+Inf")
		return
	case v+v == v && v < 0:
		printstring(w, "-Inf")
		return
	}

	const n = 7 // digits printed
	var buf [n + 7]byte
	buf[0] = '+'
	e := 0 // exp
	if v == 0 {
		if 1/v < 0 {
			buf[0] = '-'
		}
	} else {
		if v < 0 {
			v = -v
			buf[0] = '-'
		}

		// normalize
		for v >= 10 {
			e++
			v /= 10
		}
		for v < 1 {
			e--
			v *= 10
		}

		// round
		h := 5.0
		for i := 0; i < n; i++ {
			h /= 10
		}
		v += h
		if v >= 10 {
			e++
			v /= 10
		}
	}

	// format +d.dddd+edd
	for i := 0; i < n; i++ {
		s := int(v)
		buf[i+2] = byte(s + '0')
		v -= float64(s)
		v *= 10
	}
	buf[1] = buf[2]
	buf[2] = '.'

	buf[n+2] = 'e'
	buf[n+3] = '+'
	if e < 0 {
		e = -e
		buf[n+3] = '-'
	}

	buf[n+4] = byte(e/100) + '0'
	buf[n+5] = byte(e/10)%10 + '0'
	buf[n+6] = byte(e%10) + '0'
	writeWrapper(w, buf[:])
}

func printcomplex(w io.Writer, c complex128) {
	kprint("(", real(c), imag(c), "i)")
}

func printuint(w io.Writer, v uint64) {
	var buf [100]byte
	i := len(buf)
	for i--; i > 0; i-- {
		buf[i] = byte(v%10 + '0')
		if v < 10 {
			break
		}
		v /= 10
	}
	writeWrapper(w, buf[i:])
}

func printint(w io.Writer, v int64) {
	if v < 0 {
		printstring(w, "-")
		v = -v
	}
	printuint(w, uint64(v))
}

var minhexdigits = 0 // protected by printlock

func printhex(w io.Writer, v uint64) {
	const dig = "0123456789abcdef"
	var buf [100]byte
	i := len(buf)
	for i--; i > 0; i-- {
		buf[i] = dig[v%16]
		if v < 16 && len(buf)-i >= minhexdigits {
			break
		}
		v /= 16
	}
	i--
	buf[i] = 'x'
	i--
	buf[i] = '0'
	writeWrapper(w, buf[i:])
}

func printpointer(w io.Writer, p unsafe.Pointer) {
	printhex(w, uint64(uintptr(p)))
}
func printuintptr(w io.Writer, p uintptr) {
	printhex(w, uint64(p))
}

func printstring(w io.Writer, s string) {
	writeWrapper(w, bytes(s))
}

//func printslice(s []byte) {
//	sp := (*slice)(unsafe.Pointer(&s))
//	kprint("[", len(s), "/", cap(s), "]")
//	printpointer(sp.array)
//}

func kFprint(w io.Writer, args ...interface{}) {
	for _, v := range args {
		switch t := v.(type) {
		case bool:
			printbool(w, t)
		case string:
			printstring(w, t)
		case uintptr:
			printuintptr(w, t)
		case unsafe.Pointer:
			printpointer(w, t)
		case byte:
			printuint(w, uint64(t))
		case uint:
			printuint(w, uint64(t))
		case uint32:
			printuint(w, uint64(t))
		case uint64:
			printuint(w, uint64(t))
		case int:
			printint(w, int64(t))
		case int32:
			printint(w, int64(t))
		case int64:
			printint(w, t)
		case float64:
			printfloat(w, t)
		case complex128:
			printcomplex(w, t)
		default:
			// TODO: Maybe print type if that is possible without memory allocation?
			printstring(defaultErrorWriter, "<Unknown Type>")
		}
	}
}

func kprint(args ...interface{}) {
	kFprint(defaultScreenWriter, args...)
	kFprint(defaultLogWriter, args...)
}

func kprintln(args ...interface{}) {
	kprint(args...)
	kprint("\n")
}

func kerror(args ...interface{}) {
	kFprint(defaultErrorWriter, args...)
	kFprint(defaultLogWriter, args...)
}

func kerrorln(args ...interface{}) {
	kerror(args...)
	kerror("\n")
}

func kdebug(args ...interface{}) {
	kFprint(defaultLogWriter, args...)
}

func kdebugln(args ...interface{}) {
	kdebug(args...)
	kdebug("\n")
}

func writeWrapper(w io.Writer, p []byte) {
	if w == nil {
		w = defaultScreenWriter
	}
	w.Write(*(*[]byte)(noEscape(unsafe.Pointer(&p))))
}

func bytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}))
}

// noEscape hides a pointer from escape analysis. This function is copied over
// from runtime/stubs.go
//go:nosplit
func noEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
