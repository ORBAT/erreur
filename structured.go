package erreur

import (
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Structured holds an error (and a possible cause for it) and zap Fields that provide context for
// that error.
type Structured struct {
	causer error
	err    error
	fields []zap.Field
}

// Structure returns a structured error with the given error as cause and the zap fields added as
// context. Good for adding context fields to non-structured errors.
// Returns nil if err is nil
func Structure(cause error, fields ...zap.Field) error {
	if cause == nil {
		return nil
	}
	return Structured{causer: cause, fields: fields}
}

// New returns a new structured error with the given message and fields
func New(message string, fields ...zap.Field) error {
	return Structured{err: String(message), fields: fields}
}

// Wrap cause with a new message and add context fields. Returns nil if cause is nil
func Wrap(cause error, message string, fields ...zap.Field) error {
	if cause == nil {
		return nil
	}
	return Structured{causer: cause, err: String(message), fields: fields}
}

// JSONBuffer returns a go.uber.org/zap/buffer with the JSON serialization of s
func (s Structured) JSONBuffer() *buffer.Buffer {
	// NOTE: ignoring the error here is safe with the current version of zap's JSON encoder, as it
	// is always nil
	buf, _ := zapcore.NewJSONEncoder(jsonEncConf).EncodeEntry(s.entry())
	return buf
}

// MarshalJSON implements json.Marshaler
func (s Structured) MarshalJSON() ([]byte, error) {
	buf := s.JSONBuffer()
	bufBs := buf.Bytes()
	bs := make([]byte, len(bufBs))
	copy(bs, bufBs)
	buf.Free()

	return bs, nil
}
// JSON turns s into a JSON string. Shortcut for MarshalJSON that ignores the returned error (uses
// zap's JSON encoder under the hood which never returns errors, so this is safe™)
func (s Structured) JSON() string {
	bs, _ := s.MarshalJSON()
	return string(bs)
}

// Fields returns the fields of s and its causes (recursively)
func (s Structured) Fields() []zapcore.Field {
	// reserve space for our fields and a potential cause object
	fs := make([]zapcore.Field, 0, len(s.fields)+1)

	fs = append(fs, s.fields...)

	if cause := s.Unwrap(); cause != nil {
		if stre, ok := cause.(Structured); ok {
			fs = append(fs, zap.Object("cause", stre))
		}
		return fs
	}

	if stre, ok := AsStructured(s.Unwrap()); ok {
		fs = append(fs, zap.Object("cause", stre))
		return fs
	}

	return fs
}

// MarshalLogObject implements zapcore.ObjectMarshaler. This means that you can do the following:
//   zap.Object("error", s)
//
// See Field for a convenience function
func (s Structured) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString("msg", s.errorOrCause())
	for _, field := range s.Fields() {
		field.AddTo(oe)
	}
	return nil
}

// Unwrap returns the cause of this error, or nil if there is none. Implements the new experimental
// Unwrap interface in https://golang.org/x/exp/errors
func (s Structured) Unwrap() error {
	return s.causer
}

// Cause is the same as Unwrap, but implements the interface in https://github.com/pkg/errors
func (s Structured) Cause() error {
	return s.causer
}

func (s Structured) errorOrCause() string {
	if s.err != nil {
		return s.err.Error()
	}
	return s.causer.Error()
}

// Error returns just the message of s, with no context fields
func (s Structured) Error() string {
	if s.err != nil {
		if s.causer == nil { // only an error but no cause, so return that
			return s.err.Error()
		} else { // have an error and a cause for it, return both
			return s.err.Error() + ": " + s.causer.Error()
		}
	}

	// just a cause, so created with Structure()
	return s.causer.Error()
}

func (s Structured) entry() (zapcore.Entry, []zapcore.Field) {
	return zapcore.Entry{Message: s.errorOrCause()}, s.Fields()
}

// AsStructured is a shortcut for extracting a structured error from e's error chain. If ok is
// false, no matching error was found.
func AsStructured(e error) (err Structured, ok bool) {
	var s Structured
	for {
		if stre, ok := e.(Structured); ok {
			s = stre
			break
		}

		cause, ok := e.(wrapper)
		if !ok {
			break
		}
		e = cause.Unwrap()
	}

	return s, s.causer != nil || s.err != nil
}

// IsStructured returns true if e or any error in its cause chain is a Structured.
// Shortcut for
//   _, found := AsStructured(e)
func IsStructured(e error) bool {
	_, found := AsStructured(e)
	return found
}

// Field returns a zap field for err under the key "error". If err is nil, returns a no-op field. If
// err is a structured error or has one in its error chain, returns a zap.Object field, if err is a
// plain 'ol error, returns zap.Error
func Field(err error) zapcore.Field {
	if err == nil {
		return zap.Skip()
	}
	stre, ok := AsStructured(err)
	if ok {
		return zap.Object("error", stre)
	}
	return zap.Error(err)
}

// String is a lightweight string-based error. It has no "constructor", so type conversion should be
// used instead:
// 	const err = erreur.String("some error text")
type String string

// Error implements the error interface
func (es String) Error() string {
	return string(es)
}

type wrapper interface {
	Unwrap() error
}

var jsonEncConf zapcore.EncoderConfig

func init() {
	jsonEncConf = zap.NewProductionEncoderConfig()
	jsonEncConf.CallerKey = ""
	jsonEncConf.StacktraceKey = ""
	jsonEncConf.LevelKey = ""
	jsonEncConf.TimeKey = ""
	jsonEncConf.NameKey = ""
	jsonEncConf.EncodeCaller = nil
}
