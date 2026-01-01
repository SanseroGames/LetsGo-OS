package log

import (
	"io"
	"reflect"
	"unsafe"
)

var (
	defaultLogWriters   []io.Writer
	defaultDebugWriters []io.Writer
	defaultErrorWriters []io.Writer
)

// Copied and adapted code from go runtime
func printBool(w io.Writer, v bool) {
	if v {
		printString(w, "true")
	} else {
		printString(w, "false")
	}
}

func printFloat(w io.Writer, v float64) {
	switch {
	case v != v:
		printString(w, "NaN")
		return
	case v+v == v && v > 0:
		printString(w, "+Inf")
		return
	case v+v == v && v < 0:
		printString(w, "-Inf")
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

func printComplex(w io.Writer, c complex128) {
	KWPrint(w, "(", real(c), imag(c), "i)")
}

func printUint(w io.Writer, v uint64) {
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

func printInt(w io.Writer, v int64) {
	if v < 0 {
		printString(w, "-")
		v = -v
	}
	printUint(w, uint64(v))
}

func printHex(w io.Writer, v uint64) {
	const dig = "0123456789abcdef"
	var buf [100]byte
	i := len(buf)
	for i--; i > 0; i-- {
		buf[i] = dig[v%16]
		if v < 16 && len(buf)-i >= 0 {
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

func printPointer(w io.Writer, p unsafe.Pointer) {
	printHex(w, uint64(uintptr(p)))
}

func printUintptr(w io.Writer, p uintptr) {
	printHex(w, uint64(p))
}

func printString(w io.Writer, s string) {
	writeWrapper(w, bytes(s))
}

func printAny(w io.Writer, value reflect.Value) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		printInt(w, value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		printUint(w, value.Uint())
	case reflect.Bool:
		printBool(w, value.Bool())
	case reflect.Uintptr:
		printUintptr(w, uintptr(value.Uint()))
	case reflect.String:
		printString(w, "\"")
		printString(w, value.String())
		printString(w, "\"")
	case reflect.Array, reflect.Slice:
		printAnySlice(w, value)
	default:
		printAnyObject(w, value)
	}
}

func printAnyObject(w io.Writer, value reflect.Value) {
	printString(w, value.Type().String())
	if value.Kind() == reflect.Pointer {
		printString(w, " (at ")
		printPointer(w, value.UnsafePointer())
		printString(w, ")")
	}
}

func printAnySlice(w io.Writer, value reflect.Value) {
	printString(w, value.Type().String())
	printString(w, "[")
	for i := range value.Len() {
		elem := value.Index(i)
		printAny(w, elem)
		if i != value.Len()-1 {
			printString(w, ",")
		}
	}
	printString(w, "]")
}

func KWPrint(w io.Writer, args ...any) {
	// basic types are covered here, more complex types are done with reflection in printAny.
	for _, v := range args {
		switch t := v.(type) {
		case bool:
			printBool(w, t)
		case string:
			printString(w, t)
		case uintptr:
			printUintptr(w, t)
		case unsafe.Pointer:
			printPointer(w, t)
		case byte:
			printUint(w, uint64(t))
		case uint:
			printUint(w, uint64(t))
		case uint32:
			printUint(w, uint64(t))
		case uint64:
			printUint(w, uint64(t))
		case int:
			printInt(w, int64(t))
		case int32:
			printInt(w, int64(t))
		case int64:
			printInt(w, t)
		case float64:
			printFloat(w, t)
		case complex128:
			printComplex(w, t)
		default:
			// values printed here are considered debug values and are formatted as such
			value := reflect.ValueOf(v)
			printAny(w, value)
		}
	}
}

func KPrint(args ...any) {
	for _, w := range defaultLogWriters {
		KWPrint(w, args...)
	}
}

func KPrintLn(args ...any) {
	KPrint(args...)
	KPrint("\n")
}

func KError(args ...any) {
	for _, w := range defaultErrorWriters {
		KWPrint(w, args...)
	}
}

func KErrorLn(args ...any) {
	KError(args...)
	KError("\n")
}

func KDebug(args ...any) {
	for _, w := range defaultDebugWriters {
		KWPrint(w, args...)
	}
}

func KDebugLn(args ...any) {
	KDebug(args...)
	KDebug("\n")
}

func writeWrapper(w io.Writer, p []byte) {
	if w == nil {
		return
		// w = defaultScreenWritker
	}
	w.Write(*(*[]byte)(noEscape(unsafe.Pointer(&p))))
}

func bytes(s string) []byte {
	stringData := unsafe.StringData(s)
	return unsafe.Slice(stringData, len(s))
}

// noEscape hides a pointer from escape analysis. This function is copied over
// from runtime/stubs.go
//
//go:nosplit
func noEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func SetDefaultLogWriters(writers []io.Writer) {
	defaultLogWriters = writers
}

func SetDefaultDebugLogWriters(writers []io.Writer) {
	defaultDebugWriters = writers
}

func SetDefaultErrorLogWriters(writers []io.Writer) {
	defaultErrorWriters = writers
}
