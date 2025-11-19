package pkg

const (
	// GrpcPriority 如果停止需要有顺序依赖通过这里自定义priority处理 [EN] GrpcPriority If the stop needs to have order dependencies, pass this custom priority processing
	GrpcPriority      = -1
	DefaultPriority   = 1
	AsyncTaskPriority = 2
	DBPriority        = 3
)

type LoaderService interface {
	LoadType() string
	Load() error
	StopPriority() int
	Stop() error
}
