package comm

import (
	"fmt"
	"os"
	"os/signal"
)

func EXIT() { os.Exit(0) }

func WaitSignalKill() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
}

var LF = []byte{'\n'}

func fn_log(_, format string, v ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, v...)
	_, _ = os.Stdout.Write(LF)
}

func LOGI(format string, v ...any) { fn_log("I", format, v...) }
func LOGD(format string, v ...any) { fn_log("D", format, v...) }
func LOGW(format string, v ...any) { fn_log("W", format, v...) }
func LOGE(format string, v ...any) { fn_log("E", format, v...) }
func FAIL(format string, v ...any) {
	fn_log("F", format, v...)
	os.Exit(1)
}
