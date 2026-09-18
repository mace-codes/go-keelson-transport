package logger

// Noop returns a Logger that discards every entry. It is the Transport's
// default, so a consumer who supplies no logger gets silence rather than an
// imposed default. Adapters also return it when handed a nil logger, which
// keeps call sites free of nil checks.
func Noop() Logger {
	return noop{}
}

type noop struct{}

func (noop) Debug(string, ...Field) {}
func (noop) Info(string, ...Field)  {}
func (noop) Warn(string, ...Field)  {}
func (noop) Error(string, ...Field) {}

func (n noop) With(...Field) Logger { return n }
