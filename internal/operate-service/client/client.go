package client

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
)

type IClient interface {
	// System custom configuration
	CreateSystemCustom(ctx context.Context, userID, orgID string, key orm.SystemCustomKey, mode orm.SystemCustomMode, custom orm.SystemCustom) *err_code.Status
	GetSystemCustom(ctx context.Context, mode orm.SystemCustomMode) (*orm.SystemCustom, *err_code.Status)

	AddClientRecord(ctx context.Context, clientId string) *err_code.Status
	GetClientStatistic(ctx context.Context, startDate, endDate string) (*orm.ClientStatistic, *err_code.Status)
}
