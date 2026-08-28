package service

import (
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/gin-gonic/gin"
)

func UploadCustomTab(ctx *gin.Context, userId, orgId, mode string, req *request.CustomTabConfig) error {
	_, err := operate.CreateSystemCustomTab(ctx.Request.Context(), &operate_service.CreateSystemCustomTabReq{
		OrgId:  orgId,
		UserId: userId,
		Mode:   mode,
		Tab:    &operate_service.Tab{TabLogoPath: req.TabLogo.Key, TabTitle: req.TabTitle},
	})
	if err != nil {
		return err
	}
	return nil
}

func UploadCustomLogin(ctx *gin.Context, userId, orgId, mode string, req *request.CustomLoginConfig) error {
	_, err := operate.CreateSystemCustomLogin(ctx.Request.Context(), &operate_service.CreateSystemCustomLoginReq{
		OrgId:  orgId,
		UserId: userId,
		Mode:   mode,
		Login: &operate_service.Login{
			LoginBgPath:      req.LoginBg.Key,
			LoginLogo:        req.LoginLogo.Key,
			LoginButtonColor: req.LoginButtonColor,
			LoginWelcomeText: req.LoginWelcomeText},
	})
	if err != nil {
		return err
	}
	return nil
}

func UploadCustomHome(ctx *gin.Context, userId, orgId, mode string, req *request.CustomHomeConfig) error {
	_, err := operate.CreateSystemCustomHome(ctx.Request.Context(), &operate_service.CreateSystemCustomHomeReq{
		OrgId:  orgId,
		UserId: userId,
		Mode:   mode,
		Home: &operate_service.Home{HomeBgColor: req.HomeBgColor,
			HomeLogoPath: req.HomeLogo.Key,
			HomeName:     req.HomeName},
	})
	if err != nil {
		return err
	}
	return nil
}

func UploadCustomGeneralAgent(ctx *gin.Context, userId, orgId, mode string, req *request.CustomGeneralAgentConfig) error {
	_, err := operate.CreateSystemCustomGeneralAgent(ctx.Request.Context(), &operate_service.CreateSystemCustomGeneralAgentReq{
		OrgId:  orgId,
		UserId: userId,
		Mode:   mode,
		GeneralAgent: &operate_service.GeneralAgent{
			GeneralAgentIcon:     req.GeneralAgentIcon.Key,
			GeneralAgentWelcome:  req.GeneralAgentWelcome,
			GeneralAgentMenuName: req.GeneralAgentMenuName,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
