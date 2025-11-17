package oti

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

type logEntry struct {
	level int
	msg   string
	kv    []any
	err   error
	isErr bool
}

type testSink struct {
	mu     sync.Mutex
	values []any
	// shared slice pointer so all derived sinks append to the same backing slice
	entries *[]logEntry
}

type StringerImp int

func (si StringerImp) String() string { return fmt.Sprintf("Stringer:%d", si) }

type GoStringerImpl int

func (si GoStringerImpl) String() string { return fmt.Sprintf("GoStringer:%d", si) }

type StringBase string

// Implement logr.LogSink
func (s *testSink) Init(_ logr.RuntimeInfo) {}
func (s *testSink) Enabled(level int) bool  { return true }
func (s *testSink) Info(level int, msg string, kvList ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	combined := append([]any{}, s.values...)
	combined = append(combined, kvList...)
	*s.entries = append(*s.entries, logEntry{level: level, msg: msg, kv: combined})
}
func (s *testSink) Error(err error, msg string, kvList ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	combined := append([]any{}, s.values...)
	combined = append(combined, kvList...)
	*s.entries = append(*s.entries, logEntry{msg: msg, kv: combined, err: err, isErr: true})
}
func (s *testSink) WithValues(kvList ...any) logr.LogSink {
	return &testSink{values: append(append([]any{}, s.values...), kvList...), entries: s.entries}
}
func (s *testSink) WithName(_ string) logr.LogSink { return s }

// TestLogTableDriven validates Log and LogError using a strict table where the inputs are
// (level, msg, args) and we assert the resulting single log entry against the expected logEntry.
func TestLogTableDriven(t *testing.T) {
	type testCase struct {
		name   string
		level  int
		msg    string
		args   []any
		expect logEntry
		useErr bool // derived from expect.isErr (helper flag for clarity)
	}

	cases := []testCase{
		{
			name:  "simple info no sanitization needed",
			level: 1,
			msg:   "hello",
			args:  []any{"k1", "v1"},
			expect: logEntry{
				level: 1,
				msg:   "hello",
				kv:    []any{"k1", "v1"}, // no dots to change in key
			},
		},
		{
			name:  "custom string alias key is sanitized",
			level: 0,
			msg:   "alias",
			args:  func() []any { type myKey string; return []any{myKey("alpha.beta"), "val"} }(),
			expect: logEntry{
				level: 0,
				msg:   "alias",
				kv:    []any{"alpha_beta", "val"},
			},
		},
		{
			name:  "key sanitization (dots to underscores on key positions)",
			level: 0,
			msg:   "pair test",
			args:  []any{"a.b", "value.one", "c.d", "value.two"},
			expect: logEntry{
				level: 0,
				msg:   "pair test",
				// keys sanitized, values untouched
				kv: []any{"a_b", "value.one", "c_d", "value.two"},
			},
		},
		{
			name:  "multiple pairs mixed key sanitization only",
			level: 2,
			msg:   "multi",
			args:  []any{"alpha.beta", "one.two.three", "plain", "nochange"},
			expect: logEntry{
				level: 2,
				msg:   "multi",
				kv:    []any{"alpha_beta", "one.two.three", "plain", "nochange"},
			},
		},
		{
			name:  "type sanitization",
			level: 0,
			msg:   "pair test",
			args: []any{
				attribute.Key("a.b"), "value.one",
				StringerImp(2), "value.two",
				GoStringerImpl(3), "value.three",
				StringBase("String.Base"), "value.four",
				42, "value.five",
				nil, "value.six",
			},
			expect: logEntry{
				level: 0,
				msg:   "pair test",
				// keys sanitized, values untouched
				kv: []any{
					"a_b", "value.one",
					"Stringer:2", "value.two",
					"GoStringer:3", "value.three",
					"String_Base", "value.four",
					"int:42", "value.five",
					"<nil>", "value.six",
				},
			},
		},
		{
			name:   "nil key",
			level:  0,
			msg:    "nil case",
			args:   []any{nil, "value"},
			expect: logEntry{level: 0, msg: "nil case", kv: []any{"<nil>", "value"}},
		},
		{
			name:   "typed nil pointer key",
			level:  0,
			msg:    "typed nil ptr",
			args:   func() []any { var sp *StringBase = nil; return []any{sp, "val"} }(),
			expect: logEntry{level: 0, msg: "typed nil ptr", kv: []any{"*oti_StringBase:<nil>", "val"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shared := []logEntry{}
			sink := &testSink{entries: &shared}
			logger := logr.New(sink)
			ctx := logr.NewContext(context.Background(), logger)

			if tc.useErr {
				// clone error message for comparison (don't reuse same pointer if mutated elsewhere)
				LogError(ctx, errors.New(tc.expect.err.Error()), tc.msg, tc.args...)
			} else {
				Log(ctx, tc.level, tc.msg, tc.args...)
			}

			require.Len(t, *sink.entries, 1, "expected exactly one log entry recorded")
			got := (*sink.entries)[0]

			// Compare expected fields; for error we match message.
			assert.Equal(t, tc.expect.msg, got.msg, "message mismatch")
			if !tc.useErr { // level only meaningful for info logs
				assert.Equal(t, tc.expect.level, got.level, "level mismatch")
			}
			assert.True(t, reflect.DeepEqual(tc.expect.kv, got.kv), "kv mismatch expected=%v got=%v", tc.expect.kv, got.kv)
			assert.Equal(t, tc.expect.isErr, got.isErr, "isErr flag mismatch")
			if tc.expect.isErr {
				require.NotNil(t, got.err, "expected an error object")
				assert.EqualError(t, got.err, tc.expect.err.Error())
			} else {
				assert.Nil(t, got.err)
			}
		})
	}
}

func TestDotToUsHelpers(t *testing.T) {
	old := ReplaceDotToUs
	defer func() { ReplaceDotToUs = old }()
	ReplaceDotToUs = true
	assert.Equal(t, "a_b", DotToUs("a.b"))
	assert.Equal(t, "keep.me", DotToUsIf("keep.me", false))
	assert.Equal(t, "keep_me", DotToUsIf("keep.me", true))
}
