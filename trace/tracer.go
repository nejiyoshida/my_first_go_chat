package trace

import (
	"fmt"
	"io"
)

// 必要な時に必要な内容を利用者が実装
type Tracer interface {
	Trace(...interface{}) // 可変長引数かつinterfac{}を受けるので、何でもあり
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	t.out.Write([]byte(fmt.Sprint(a...)))
	t.out.Write([]byte("\n"))
}

func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

type nilTracer struct{}

func (t *nilTracer) Trace(a ...interface{}) {}

func Off() Tracer {
	return &nilTracer{}
}
