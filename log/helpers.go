package log

func String(key, val string) Field {
	return Field{Key: key, Encoder: stringVal(val)}
}

func Int(key string, val int) Field {
	return Field{Key: key, Encoder: intVal(val)}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Encoder: boolVal(val)}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Encoder: floatVal(val)}
}

func StringSlice(key string, val []string) Field {
	return Field{Key: key, Encoder: stringSliceVal(val)}
}

func Map(key string, val map[string]any) Field {
	return Field{Key: key, Encoder: mapStringAny(val)}
}

func Err(key string, val error) Field {
	return Field{Key: key, Encoder: errorVal{err: val}}
}
