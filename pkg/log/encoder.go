package log

import (
	"sort"
	"strconv"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type noDuplicateEncoder struct {
	zapcore.Encoder

	reserved map[string]bool
}

func newNoDuplicateEncoder(enc zapcore.Encoder, reserved ...string) zapcore.Encoder {
	rmap := make(map[string]bool, len(reserved))
	for _, r := range reserved {
		rmap[r] = true
	}
	return &noDuplicateEncoder{Encoder: enc, reserved: rmap}
}

func (n *noDuplicateEncoder) Clone() zapcore.Encoder {
	return &noDuplicateEncoder{Encoder: n.Encoder.Clone(), reserved: n.reserved}
}

func (n *noDuplicateEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	n.deduplicate(fields)
	return n.Encoder.EncodeEntry(entry, fields)
}

func (n *noDuplicateEncoder) deduplicate(fields []zapcore.Field) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Key < fields[j].Key
	})

	var prevKey string
	var curKey string
	repeat := 1
	for i := range fields {
		curKey = fields[i].Key
		if curKey == "" {
			continue
		}
		if prevKey != curKey {
			repeat = 1
		}
		if prevKey == curKey || n.reserved[curKey] {
			fields[i].Key += strconv.Itoa(repeat)
			repeat++
		}
		prevKey = curKey
	}
}
