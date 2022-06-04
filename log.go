package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	prefix    string     // prefix on each line to identify the logger (but see Lmsgprefix)
	flag      int        // properties
	out       io.Writer  // destination for output
	buf       []byte     // for accumulating text to write
	isDiscard int32      // atomic boolean: whether out == io.Discard
}

func New(out io.Writer, prefix string, flag int) *Logger {
	l := &Logger{out: out, prefix: prefix, flag: flag}
	if out == io.Discard {
		l.isDiscard = 1
	}
	return l
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	isDiscard := int32(0)
	if w == io.Discard {
		isDiscard = 1
	}
	atomic.StoreInt32(&l.isDiscard, isDiscard)
}

var std = New(os.Stderr, "", LstdFlags)

// Default returns the standard logger used by the package-level output functions.
func Default() *Logger { return std }

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank and Lmsgprefix is unset),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided),
//   * l.prefix (if it's not blank and Lmsgprefix is set).
func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	if l.flag&Lmsgprefix == 0 {
		*buf = append(*buf, l.prefix...)
	}
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
	if l.flag&Lmsgprefix != 0 {
		*buf = append(*buf, l.prefix...)
	}
}

func (l *Logger) Output(functionName, color string, err error, calldepth int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}

	if functionName == "github.com/perfectogo/log.Error" {
		if err != nil {
			l.buf = []byte(color + "TIME: " + string(l.buf) + Reset)
		} else {
			l.buf = []byte(color + "TIME: " + string(l.buf) + Reset)
		}
	} else {
		l.buf = []byte(color + "TIME: " + string(l.buf) + Reset)
	}
	_, err = l.out.Write(l.buf)
	return err

}

func getCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return fmt.Sprintf("%s", runtime.FuncForPC(pc).Name())
}

func Println(v ...any) {
	pc, filename, line, _ := runtime.Caller(1)

	fn := func(uintptr) string {
		pc, _, _, _ := runtime.Caller(1)
		return fmt.Sprintf("%s", runtime.FuncForPC(pc).Name())
	}

	fnName := getCurrentFuncName()
	color := Green
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}

	color = White
	std.Output(
		fnName, color, nil, 2,
		"\nPATH:\t"+filename+"\nFUNCTION: "+fn(pc)+"\nLOG LINE: "+strconv.Itoa(line)+color+"\nINFO: "+fmt.Sprint(v...)+Reset,
	)
}

func Info(v ...any) {
	fnName := getCurrentFuncName()
	color := Blue
	_, filename, line, _ := runtime.Caller(1)

	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.Output(
		fnName, color, nil, 2,
		"\n\tPATH: "+filename+"\n\tLOG LINE: "+strconv.Itoa(line)+"\n\tINFO: "+color+fmt.Sprint(v...)+Reset,
	)
}

func Error(msg string, err error) {
	_, filename, line, _ := runtime.Caller(1)
	fnName := getCurrentFuncName()
	color := Green
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	if err != nil {
		color = Red
		std.Output(
			fnName, color, err, 2,
			"\n\tPATH: "+filename+"\n\tLOG LINE: "+strconv.Itoa(line)+color+"\n\tMESSAGE: "+fmt.Sprint(msg)+"\n\tERROR: "+err.Error()+Reset,
		)
		return
	}
	std.Output(
		fnName, color, err, 2,
		"\n\tPATH: "+filename+"\n\tLOG LINE: "+strconv.Itoa(line)+color+"\n\tMESSAGE: "+fmt.Sprint(msg)+"\n\tERROR: NO ERROR"+Reset,
	)
}

func Warning(v ...any) {
	_, filename, line, _ := runtime.Caller(1)
	fnName := getCurrentFuncName()
	color := Yellow
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.Output(
		fnName, color, nil, 2,
		"\nWARNING LOG\n\tPATH: "+filename+"\n\tLOG LINE: "+strconv.Itoa(line)+"\n\tWARNING: "+color+fmt.Sprint(v...)+Reset)
}
