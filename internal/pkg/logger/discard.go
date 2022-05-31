package logger

// DiscardLogger ...
type DiscardLogger struct{
	
}

// Errorf do nothing, just discard resty log
func (DiscardLogger) Errorf(format string, v ...interface{}) {

}

// Warnf do nothing, just discard resty log
func (DiscardLogger) Warnf(format string, v ...interface{}) {

}

// Debugf do nothing, just discard resty log
func (DiscardLogger) Debugf(format string, v ...interface{}) {

}