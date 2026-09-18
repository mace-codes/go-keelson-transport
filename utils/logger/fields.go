package logger

import "time"

// String returns a Field holding a string value.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int returns a Field holding an int value.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 returns a Field holding an int64 value.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 returns a Field holding a float64 value.
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool returns a Field holding a bool value.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration returns a Field holding a time.Duration value.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time returns a Field holding a time.Time value.
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Err returns a Field holding an error under ErrorKey. Adapters map this onto
// their library's native error field where one exists.
func Err(err error) Field {
	return Field{Key: ErrorKey, Value: err}
}

// Any returns a Field holding a value of any type, left to the adapter to
// encode. Prefer the typed constructors where one fits.
func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}
