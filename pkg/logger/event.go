package logger

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

type event struct {
	ts    int64
	level level
	msg   string
}

func (e event) json(builder *bytes.Buffer) []byte {
	builder.Reset()

	builder.WriteString(`{"time":`)
	builder.WriteString(strconv.FormatInt(e.ts, 10))
	builder.WriteString(`,"level":"`)
	builder.WriteString(levelValues[e.level])
	builder.WriteString(`","msg":"`)
	builder.WriteString(e.msg)
	builder.WriteString(`"}`)
	builder.WriteString("\n")

	return builder.Bytes()
}

func (e event) text(builder *bytes.Buffer) []byte {
	builder.Reset()

	builder.WriteString(time.Unix(e.ts, 0).Format(time.RFC3339))
	builder.WriteString(` `)
	builder.WriteString(fmt.Sprintf("%-5s", levelValues[e.level]))
	builder.WriteString(` `)
	builder.WriteString(e.msg)
	builder.WriteString("\n")

	return builder.Bytes()
}
