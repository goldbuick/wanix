package api

import (
	"fmt"
	"time"

	"tractor.dev/toolkit-go/duplex/rpc"
	"tractor.dev/wanix/fs"
)

// asfloat64 coerces duplex/msgpack numeric args to float64.
// Whole-second timestamps often arrive as int64/uint64, not float64.
func asfloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func secstofloattime(sec float64) time.Time {
	return time.Unix(int64(sec), int64((sec-float64(int64(sec)))*1e9))
}

func (s *syscaller) chtimes(r rpc.Responder, c *rpc.Call) {
	var args []any
	c.Receive(&args)

	if len(args) < 3 {
		r.Return(fmt.Errorf("chtimes: need path, atime, mtime (got %d args)", len(args)))
		return
	}

	path, ok := args[0].(string)
	if !ok {
		r.Return(fmt.Errorf("chtimes: arg 0 is not a string: %T", args[0]))
		return
	}

	// atime and mtime are in seconds (with fractional parts)
	atimeSec, ok := asfloat64(args[1])
	if !ok {
		r.Return(fmt.Errorf("chtimes: arg 1 is not numeric: %T", args[1]))
		return
	}
	atime := secstofloattime(atimeSec)

	mtimeSec, ok := asfloat64(args[2])
	if !ok {
		r.Return(fmt.Errorf("chtimes: arg 2 is not numeric: %T", args[2]))
		return
	}
	mtime := secstofloattime(mtimeSec)

	err := fs.Chtimes(s.task.NS(), path, atime, mtime)
	if err != nil {
		r.Return(err)
		return
	}
}
