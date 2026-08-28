package pkg

const (
	// GrpcPriority 如果停止需要有顺序依赖通过这里自定义priority处理
	GrpcPriority      = -1
	DefaultPriority   = 1
	AsyncTaskPriority = 2
	DBPriority        = 3
	TracePriority     = 4
)

type LoaderService interface {
	LoadType() string
	Load() error
	StopPriority() int
	Stop() error
}
