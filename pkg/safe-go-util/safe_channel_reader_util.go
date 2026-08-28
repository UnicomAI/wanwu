package safe_go_util

import (
	"context"
)

type IteratorError[T any] struct {
	Err           error
	OutputMessage bool //是否输出errMsg
	ErrMsg        T
}

type IteratorReaderResponse[T any, R any] struct {
	Data    T
	HasData bool              //是否有数据
	Stop    bool              //是否停止
	Err     *IteratorError[R] // 错误数据
}

type IteratorReader[T any, R any] struct {
	Reader    func(ctx context.Context) IteratorReaderResponse[T, R]
	Processor func(ctx context.Context, data T, rawCh chan R) (resultList []R, err *IteratorError[R])
}

func SafeChannelReceiveByIter[T any, R any](ctx context.Context, lineIter *IteratorReader[R, T]) <-chan T {
	return SafeChannelReceiveByIterCloser(ctx, lineIter, nil)
}

func SafeChannelReceiveByIterCloser[T any, R any](ctx context.Context, lineIter *IteratorReader[R, T], bizCloser func(context.Context)) <-chan T {
	rawCh := make(chan T, 128)
	var closer = func(ctx context.Context) {
		close(rawCh)
		if bizCloser != nil {
			bizCloser(ctx)
		}
	}
	// 起一个协程安全执行数据接收的方法
	SafeGo(safeCycleReceiveFuncByIter(ctx, lineIter, rawCh, closer))
	return rawCh
}

func SafeCycleReceiveByIter[T any, R any](ctx context.Context, lineIter *IteratorReader[R, T], rawCh chan T, closer func(context.Context)) error {
	defer func() {
		if closer != nil {
			closer(ctx)
		}
	}()
	for {
		resp := safeReceiveByIter(ctx, lineIter, rawCh)
		if len(resp.ResultList) > 0 {
			for _, result := range resp.ResultList {
				rawCh <- result
			}
		}
		if resp.Skip { //需要跳过
			continue
		}
		if stop(resp) { //错误或者正常结束
			return resp.Err
		}
	}
}

func IteratorResponseStop[T any, R any]() IteratorReaderResponse[T, R] {
	return IteratorReaderResponse[T, R]{Stop: true}
}

func IteratorResponseDataStop[T any, R any](data T) IteratorReaderResponse[T, R] {
	return IteratorReaderResponse[T, R]{Stop: true, HasData: true, Data: data}
}

func IteratorResponseErr[T any, R any](err *IteratorError[R]) IteratorReaderResponse[T, R] {
	return IteratorReaderResponse[T, R]{Err: err}
}

func safeCycleReceiveFuncByIter[T any, R any](ctx context.Context, lineIter *IteratorReader[R, T], rawCh chan T, closer func(context.Context)) func() {
	return func() {
		_ = SafeCycleReceiveByIter(ctx, lineIter, rawCh, closer)
	}
}

// safeReceive 安全的channel读取，支持context取消
func safeReceiveByIter[T any, R any](ctx context.Context, lineIter *IteratorReader[R, T], rawCh chan T) ChannelReceiveResult[T] {
	select {
	case <-ctx.Done():
		return ChannelReceiveResult[T]{Err: ctx.Err(), Stop: true}
	default:
		//迭代读取数据
		iterResp := lineIter.Reader(ctx)
		iterErr := iterResp.Err
		if iterErr != nil { //有错误信息
			return processError(iterErr)
		}
		if iterResp.Stop { //停止处理
			if iterResp.HasData { //可能停止时也获得到了数据
				return processData(ctx, lineIter, iterResp.Data, rawCh, true)
			}
			return ChannelStop[T]()
		}
		//处理数据
		return processData(ctx, lineIter, iterResp.Data, rawCh, false)
	}
}

func processError[T any](iterErr *IteratorError[T]) ChannelReceiveResult[T] {
	var resultList []T
	if iterErr.OutputMessage {
		resultList = []T{iterErr.ErrMsg}
	}
	return ChannelErr[T](resultList, iterErr.Err)
}

func processData[T any, R any](ctx context.Context, lineIter *IteratorReader[R, T], data R, rawCh chan T, stop bool) ChannelReceiveResult[T] {
	resultList, iteratorError := lineIter.Processor(ctx, data, rawCh)
	if iteratorError != nil {
		return processError(iteratorError)
	}
	if len(resultList) > 0 {
		return ChannelReceiveResult[T]{ResultList: resultList, Stop: stop}
	}
	if stop {
		return ChannelStop[T]()
	}
	return ChannelSkip[T]()
}
