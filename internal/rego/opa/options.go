package opa

import (
	"io"
	"time"

	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/metrics"
	"github.com/meta-quick/opa/topdown/builtins"
	"github.com/meta-quick/opa/topdown/cache"
	"github.com/meta-quick/opa/topdown/print"
)

// Result holds the evaluation result.
type Result struct {
	Result []byte
}

// EvalOpts define options for performing an evaluation.
type EvalOpts struct {
	Input                  *interface{}
	Metrics                metrics.Metrics
	Entrypoint             int32
	Time                   time.Time
	Seed                   io.Reader
	InterQueryBuiltinCache cache.InterQueryCache
	NDBuiltinCache         builtins.NDBCache
	PrintHook              print.Hook
	Capabilities           *ast.Capabilities
}
