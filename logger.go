package statful

type Logger func(...interface{})

var (
	emptyLogger = func(_ ...interface{}) {}
	debugLog = emptyLogger
	errorLog = emptyLogger
)

func SetDebugLogger(logger Logger) {
	debugLog = logger
}

func SetErrorLogger(logger Logger) {
	errorLog = logger
}
