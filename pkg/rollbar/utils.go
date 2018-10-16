package rollbar

import (
	"fmt"
	"runtime"
	"strings"
)

func stackTrace(skip, depth int) string {
	pc := make([]uintptr, depth)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return ``
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)
	stacktraces := make([]string, 0)

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, `runtime/`) {
			break
		}

		stacktraces = append(stacktraces, fmt.Sprintf(`%s:%d`, frame.Function, frame.Line))
		if !more {
			break
		}
	}

	if len(stacktraces) == 0 {
		return ``
	}

	if len(stacktraces) == 1 {
		return fmt.Sprintf("\nfrom %s", stacktraces[0])
	}

	return fmt.Sprintf("\nfrom\n\t- %s", strings.Join(stacktraces, "\n\t- "))
}
