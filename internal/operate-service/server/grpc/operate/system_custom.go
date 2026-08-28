package operate

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CreateSystemCustomTab(ctx context.Context, req *operate_service.CreateSystemCustomTabReq) (*emptypb.Empty, error) {
	err := s.cli.CreateSystemCustom(ctx, req.UserId, req.OrgId, orm.SystemCustomTabKey, orm.SystemCustomMode(req.Mode), toSystemCustomTab(req))
	if err != nil {
		return nil, errStatus(errs.Code_OperateCustom, err)
	}
	return nil, nil
}

func (s *Service) CreateSystemCustomLogin(ctx context.Context, req *operate_service.CreateSystemCustomLoginReq) (*emptypb.Empty, error) {
	err := s.cli.CreateSystemCustom(ctx, req.UserId, req.OrgId, orm.SystemCustomLoginKey, orm.SystemCustomMode(req.Mode), toSystemCustomLogin(req))
	if err != nil {
		return nil, errStatus(errs.Code_OperateCustom, err)
	}
	return nil, nil
}

func (s *Service) CreateSystemCustomHome(ctx context.Context, req *operate_service.CreateSystemCustomHomeReq) (*emptypb.Empty, error) {
	err := s.cli.CreateSystemCustom(ctx, req.UserId, req.OrgId, orm.SystemCustomHomeKey, orm.SystemCustomMode(req.Mode), toSystemCustomHome(req))
	if err != nil {
		return nil, errStatus(errs.Code_OperateCustom, err)
	}
	return nil, nil
}

func (s *Service) CreateSystemCustomGeneralAgent(ctx context.Context, req *operate_service.CreateSystemCustomGeneralAgentReq) (*emptypb.Empty, error) {
	err := s.cli.CreateSystemCustom(ctx, req.UserId, req.OrgId, orm.SystemCustomGeneralAgentKey, orm.SystemCustomMode(req.Mode), toSystemCustomGeneralAgent(req))
	if err != nil {
		return nil, errStatus(errs.Code_OperateCustom, err)
	}
	return nil, nil
}

func (s *Service) GetSystemCustom(ctx context.Context, req *operate_service.GetSystemCustomReq) (*operate_service.SystemCustom, error) {
	systemCustom, err := s.cli.GetSystemCustom(ctx, orm.SystemCustomMode(req.Mode))
	if err != nil {
		return nil, errStatus(errs.Code_OperateCustom, err)
	}
	return toProtoSystemCustom(systemCustom), nil
}

func toProtoSystemCustom(system *orm.SystemCustom) *operate_service.SystemCustom {
	return &operate_service.SystemCustom{
		Tab: &operate_service.Tab{
			TabLogoPath: system.Tab.LogoPath,
			TabTitle:    system.Tab.Title,
		},
		Login: &operate_service.Login{
			LoginBgPath:      system.Login.LoginBgPath,
			LoginLogo:        system.Login.LogoPath,
			LoginWelcomeText: system.Login.WelcomeText,
			LoginButtonColor: system.Login.ButtonColor},
		Home: &operate_service.Home{
			HomeLogoPath: system.Home.LogoPath,
			HomeName:     system.Home.Name,
			HomeBgColor:  system.Home.BgColor,
		},
		GeneralAgent: &operate_service.GeneralAgent{
			GeneralAgentIcon:     system.GeneralAgent.GeneralAgentIconPath,
			GeneralAgentWelcome:  system.GeneralAgent.GeneralAgentWelcome,
			GeneralAgentMenuName: system.GeneralAgent.GeneralAgentMenuName,
		},
	}

}

func toSystemCustomTab(req *operate_service.CreateSystemCustomTabReq) orm.SystemCustom {
	return orm.SystemCustom{
		Tab: orm.TabConfig{
			LogoPath: req.Tab.GetTabLogoPath(),
			Title:    req.Tab.GetTabTitle(),
		},
	}
}

func toSystemCustomLogin(req *operate_service.CreateSystemCustomLoginReq) orm.SystemCustom {
	return orm.SystemCustom{
		Login: orm.LoginConfig{
			LoginBgPath: req.Login.GetLoginBgPath(),
			LogoPath:    req.Login.GetLoginLogo(),
			WelcomeText: req.Login.GetLoginWelcomeText(),
			ButtonColor: req.Login.GetLoginButtonColor(),
		},
	}
}

func toSystemCustomHome(req *operate_service.CreateSystemCustomHomeReq) orm.SystemCustom {
	return orm.SystemCustom{
		Home: orm.HomeConfig{
			LogoPath: req.Home.GetHomeLogoPath(),
			Name:     req.Home.GetHomeName(),
			BgColor:  req.Home.GetHomeBgColor(),
		},
	}
}

func toSystemCustomGeneralAgent(req *operate_service.CreateSystemCustomGeneralAgentReq) orm.SystemCustom {
	return orm.SystemCustom{
		GeneralAgent: orm.GeneralAgentConfig{
			GeneralAgentIconPath: req.GeneralAgent.GetGeneralAgentIcon(),
			GeneralAgentWelcome:  req.GeneralAgent.GetGeneralAgentWelcome(),
			GeneralAgentMenuName: req.GeneralAgent.GetGeneralAgentMenuName(),
		},
	}
}
