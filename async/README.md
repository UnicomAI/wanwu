# Go Decentralized Async Task Framework
基于golang实现的去中心化异步任务框架

**go-async特点**

- [x] 基于DB管理异步任务，以package代码库的形式引入使用
- [x] 适应微服务集群多副本的部署模式，每个服务节点中引入的go-async之间，是去中心化的关系
- [x] 支持用户自定义不同类型异步任务的过程方法
- [x] 以多用户多任务队列的方式控制异步任务并发执行，支持用户指定异步任务所属队列，支持用户自定义任务队列与队列调度策略
- [x] 支持在异步任务排队、执行过程中，用户变更异步任务所属队列，暂停、运行、删除异步任务
- [x] 支持在异步任务排队、执行过程中，引入go-async的服务节点关闭、重启，基于此外部对于异步任务的状态无感
- [x] Goroutine Safe & Developer Friendly

## 安装

```go
go get github.com/gromitlee/go-async
```

## 说明

### 设计理念

**去中心化**
1. 去中心化，即意味着go-async可以代码库的形式被引入，可以分散在一个集群的一组服务当中形成一套go-async系统；而非必需以单个或一组服务的形式对外提供使用 
2. 一套go-async系统，依赖同一个DB与DB提供的事务，完成分布式的多任务状态管理；即将中心化、一致性的责任托管给DB，go-async本身是去中心化的

**系统边界**
1. 一套go-async系统内共用同一个DB，这是不同go-async系统间的边界，一个集群中可以同时存在多套不相干的go-async系统
2. 一套go-async系统最少可由一个节点构成，一套系统内的不同节点，相互间是平等关系
3. 一套go-async系统内任务的类型全局一致

**任务运行**
1. 一个任务只有在运行中，才会被加载到一个且只有一个节点的内存中，并且该运行中的任务，只受该节点的管理，直到该任务从内存中移除 
2. 节点会以心跳的形式定时更新其上运行着的任务在DB中的状态，当节点发生异常，无法继续维持任务的心跳时，其他节点会发现并标记这些任务运行失败 
3. 当节点正常关闭时，会暂存其上运行着的任务的上下文，标记这些任务为暂停状态并从内存中移除，其他节点会发现并接管继续运行这些任务，这要求业务层在实现各种异步任务时，需要考虑从保存的上下文中正常恢复任务 
4. 暂不考虑节点管理的运行中的任务，托管在其他节点上运行的实现方案

### API

**[go-async API](api.go)**

```go
func Init(ctx context.Context, db *gorm.DB, options ...AsyncOption) error
func Stop()

func RegisterTask(taskTyp uint32, newTask async_task.ITaskFunc) error
func CreateTask(ctx context.Context, user, group string, taskTyp uint32, taskCtx string, autoRun bool) (uint32, error)

func ChangeTaskGroup(ctx context.Context, taskID uint32, group string) error

func RunTask(ctx context.Context, taskID uint32) error
func DeleteTask(ctx context.Context, taskID uint32) error
func PauseTask(ctx context.Context, taskID uint32) error

func GetTask(ctx context.Context, taskID uint32) (*async_task.Task, error)
func GetTasks(ctx context.Context, user, group string, taskTypes []uint32, states []async_task.State, offset, limit int32) ([]*async_task.Task, error)
```

- `Init`和`Stop`方法用于用户初始化和关闭go-async系统
- `RegisterTask`和`CreateTask`方法用于用户注册和创建不同类型的异步任务
- `RunTask`、`DeleteTak`和`PauseTask`方法用于用户运行/继续运行、删除和暂停异步任务
- `ChangeTaskGroup`方法用于用户变更异步任务所属队列
- `GetTask`、`GetTasks`方法用于查询异步任务，用户可见的任务状态有
  - `StateInit`已创建
  - `StatePending`排队中
  - `StateRunning`运行中
  - `StateCanceling`取消中
  - `StatePause`暂停
  - `StateFinished`结束
  - `StateFailed`失败

**[go-async task API](pkg/async/async_task/task.go)**

```go
type IReport interface {
    Phase() (RunPhase, bool)
    Context() string
}

type ITask interface {
    Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan IReport
    Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan IReport
}
```

- `Running`方法用于用户实现完成异步任务业务逻辑的过程方法
  - `ctx`是异步任务运行上下文，由go-async系统初始化时(`Init`方法)指定，不会被go-async系统本身主动取消
  - `taskCtx`是异步任务执行所需的逻辑上下文，由创建异步任务时(`CreateTask`方法)指定，或被异步任务本身动态更新、上报后，由go-async系统回传
  - `stop`用于接收go-async系统的停止信号，可能来自于用户手动暂停(`PauseTask`方法)、删除(`DeleteTask`方法)或go-async系统关闭，`Running`方法运行中最多只会收到一次停止信号
  - `IReport`用于`Running`方法向go-async系统上报异步任务执行情况与逻辑上下文
    - `Context`方法用于向go-async系统上报当前的逻辑上下文，go-async系统负责存储
    - `Phase`方法用于向go-async系统上报异步任务的执行情况
      - `RunPhaseNormal`表示任务正常执行
      - `RunPhaseFinished`表示任务结束，`bool`表示是否删除对应任务记录，之后`IReport`channel应当被关闭并退出执行
      - `RunPhaseFailed`表示任务失败，`bool`表示是否删除对应任务记录，之后`IReport`channel应当被关闭并退出执行
      - 接收到go-async系统`stop`停止信号，任务根据自身情况上报后退出执行即可；一般是上报`RunPhaseNormal`
- `Deleting`方法用于用户实现删除(`DeleteTask`方法)异步任务后，清理业务逻辑的过程方法，与`Running`方法类似

## 示例

- 示例1：[向量点积异步任务](examples/task_dot_test.go)
- 示例2：[矩阵相乘异步任务(基于向量点积并行计算)](examples/task_mm_test.go)

## TODO

- [x] 开放自定义log组件API
- [x] 开放自定义任务队列组件API(可由用户自定义队列调度策略)
- [ ] 加速任务并发启动