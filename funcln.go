package log

import (
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"
)

var (
	filename string
	line     int
	pc       uintptr
)

func Println(v ...any) {
	pc, filename, line, _ = runtime.Caller(1)
	color := Green
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	color = White
	std.Output(
		color, nil, 2,
		color+
			"\nPATH:       "+filename+
			"\nFUNCTION:   "+Fn(pc)+
			"\nLOG LINE:   "+strconv.Itoa(line)+
			"\nPRINTLN:    "+fmt.Sprint(v...)+
			Reset,
	)
}

func Infoln(v ...any) {
	color := Blue
	_, filename, line, _ := runtime.Caller(1)

	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.Output(
		color, nil, 2,
		color+
			"\nPATH:       "+filename+
			"\nFUNCTION:   "+Fn(pc)+
			"\nLOG LINE:   "+strconv.Itoa(line)+
			"\nINFO:       "+fmt.Sprint(v...)+
			Reset,
	)
}

func Errorln(msg string, err error) {
	_, filename, line, _ := runtime.Caller(1)
	color := Green
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	if err != nil {
		color = Red
		std.Output(
			color, err, 2,
			color+
				"\nPATH:       "+filename+
				"\nFUNCTION:   "+Fn(pc)+
				"\nLOG LINE:   "+strconv.Itoa(line)+color+
				"\nMESSAGE:    "+fmt.Sprint(msg)+
				"\nERROR:      "+err.Error()+
				Reset,
		)
		return
	}
	std.Output(
		color, err, 2,
		color+
			"\nPATH:       "+filename+
			"\nFUNCTION:   "+Fn(pc)+
			"\nLOG LINE:   "+strconv.Itoa(line)+
			"\nMESSAGE:    "+fmt.Sprint(msg)+
			"\nERROR:      NO ERROR"+
			Reset,
	)
}

func Warning(v ...any) {
	pc, filename, line, _ := runtime.Caller(1)
	color := Yellow
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.Output(
		color, nil, 2,
		color+
			"\nPATH:       "+filename+
			"\nFUNCTION:   "+Fn(pc)+
			"\nLOG LINE:   "+strconv.Itoa(line)+
			"\nWARNING:    "+fmt.Sprint(v...)+
			Reset)
}
