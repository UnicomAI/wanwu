package async_config

type Logger interface {
	Debugf(fmt string, i ...interface{})
	Infof(fmt string, i ...interface{})
	Warnf(fmt string, i ...interface{})
	Errorf(fmt string, i ...interface{})
}
