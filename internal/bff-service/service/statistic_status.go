package service

import (
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcErrorToHTTPStatus 从 gRPC 错误中提取状态码与错误消息，用于应用/模型统计记录。
// err == nil 时返回 (SuccessStatusCode, "")。
// 否则按 gRPC code 映射到对应的 HTTP 语义码（httpStatusFromGRPCCode 保证非零，未知 code 兜底 500），msg 取 err.Error()。
func GrpcErrorToHTTPStatus(err error) (int64, string) {
	if err == nil {
		return statistic.SuccessStatusCode, ""
	}
	return int64(httpStatusFromGRPCCode(status.Code(err))), err.Error()
}

// httpStatusFromGRPCCode 将 gRPC code 映射为 HTTP 语义状态码。
func httpStatusFromGRPCCode(c codes.Code) int {
	switch c {
	case codes.OK:
		return int(statistic.SuccessStatusCode)
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return 500
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return 400
	case codes.DeadlineExceeded:
		return 504
	case codes.NotFound:
		return 404
	case codes.AlreadyExists, codes.Aborted:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.ResourceExhausted:
		return 429
	case codes.Unimplemented:
		return 501
	case codes.Internal, codes.DataLoss:
		return 500
	case codes.Unavailable:
		return 503
	case codes.Unauthenticated:
		return 401
	default:
		return 500
	}
}

// appStreamStatisticStatus 流式统计状态码：优先用 err 精确映射（GrpcErrorToHTTPStatus），
// 否则用 SSE error: 文本兜底 500，成功返回 SuccessStatusCode。
func appStreamStatisticStatus(err error, errMsg string) (int64, string) {
	if err != nil {
		return GrpcErrorToHTTPStatus(err)
	}
	if errMsg != "" {
		return 500, errMsg
	}
	return statistic.SuccessStatusCode, ""
}

// statisticExportSuccessLabel 导出用调用结果文案（避免 Excel 显示 true/false）。
func statisticExportSuccessLabel(ok bool) string {
	if ok {
		return "成功"
	}
	return "失败"
}
