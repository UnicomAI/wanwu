package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	bff_rsautil "github.com/UnicomAI/wanwu/internal/bff-service/pkg/rsa-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	rsautil "github.com/UnicomAI/wanwu/pkg/rsa-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// --- user excel import constants ---
const (
	ExcelHeaderUserName      = "用户名"
	ExcelHeaderPassword      = "密码"
	ExcelHeaderPhone         = "电话"
	ExcelHeaderRole          = "角色"
	ExcelHeaderEmail         = "邮箱"
	MaxBatchCreateUsersLimit = 500
)

var requiredUserExcelHeaders = []string{
	ExcelHeaderUserName,
	ExcelHeaderPassword,
	ExcelHeaderPhone,
	ExcelHeaderRole,
}

func CreateUser(ctx *gin.Context, creatorID, orgID string, userCreate *request.UserCreate) (*response.UserID, error) {
	password, err := decryptCipherRSA(ctx.Request.Context(), userCreate.Cipher, userCreate.KeyID, challengeConsume)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}
	if config.Cfg().CustomInfo.UserPhoneRequired != 0 && userCreate.Phone == "" {
		return nil, fmt.Errorf("phone is empty")
	}
	if userCreate.Phone != "" {
		if err := validatePhone(userCreate.Phone); err != nil {
			return nil, fmt.Errorf("phone %s is invalid", userCreate.Phone)
		}
	}
	if userCreate.Email != "" {
		if err := validateEmail(userCreate.Email); err != nil {
			return nil, fmt.Errorf("email %s is invalid", userCreate.Email)
		}
	}
	resp, err := iam.CreateUser(ctx.Request.Context(), &iam_service.CreateUserReq{
		CreatorId: creatorID,
		OrgId:     orgID,
		UserName:  userCreate.UserName,
		Phone:     userCreate.Phone,
		Email:     userCreate.Email,
		Password:  password,
		RoleIds:   userCreate.RoleIDs,
	})
	if err != nil {
		return nil, err
	}
	return &response.UserID{UserID: resp.Id}, nil
}

func ChangeUser(ctx *gin.Context, orgID string, userUpdate *request.UserUpdate) error {
	if config.Cfg().CustomInfo.UserPhoneRequired != 0 && userUpdate.Phone == "" {
		return fmt.Errorf("phone is empty")
	}
	if userUpdate.Phone != "" {
		if err := validatePhone(userUpdate.Phone); err != nil {
			return fmt.Errorf("phone %s is invalid", userUpdate.Phone)
		}
	}
	if userUpdate.Email != "" {
		if err := validateEmail(userUpdate.Email); err != nil {
			return fmt.Errorf("email %s is invalid", userUpdate.Email)
		}
	}
	_, err := iam.UpdateUser(ctx.Request.Context(), &iam_service.UpdateUserReq{
		UserId:   userUpdate.UserID,
		OrgId:    orgID,
		UserName: userUpdate.UserName,
		Phone:    userUpdate.Phone,
		Email:    userUpdate.Email,
		RoleIds:  userUpdate.RoleIDs,
	})
	return err
}

func DeleteUser(ctx *gin.Context, userID string) error {
	_, err := iam.DeleteUser(ctx.Request.Context(), &iam_service.DeleteUserReq{
		UserId: userID,
	})
	return err
}

func GetUserInfo(ctx *gin.Context, userID, orgID string) (*response.UserInfo, error) {
	resp, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	return toUserInfo(ctx, resp), nil
}

func GetUserList(ctx *gin.Context, orgID, name string, roleIDs []string, pageNo, pageSize int32) (*response.PageResult, error) {
	resp, err := iam.GetUserList(ctx.Request.Context(), &iam_service.GetUserListReq{
		OrgId:    orgID,
		UserName: name,
		Email:    name,
		RoleIds:  roleIDs,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	var users []*response.UserInfo
	for _, user := range resp.Users {
		users = append(users, toUserInfo(ctx, user))
	}
	return &response.PageResult{
		List:     users,
		Total:    resp.Total,
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

func GetUserListByUserIds(ctx *gin.Context, userIDs []string) (*response.ListResult, error) {
	resp, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, err
	}
	var users []*response.IDName
	for _, user := range resp.Selects {
		users = append(users, &response.IDName{
			ID:   user.Id,
			Name: user.Name,
		})
	}
	return &response.ListResult{List: users, Total: int64(len(users))}, nil
}

func GetUsersByOrgIDs(ctx context.Context, userID string, req *request.OrgIDsReq) (*response.Users, error) {
	resp, err := iam.GetUsersByOrgIDs(ctx, &iam_service.GetUsersByOrgIDsReq{
		OrgIds:   req.OrgIDList,
		IsAllOrg: req.IsAllOrg,
		UserId:   userID,
	})
	if err != nil {
		return nil, err
	}
	var users []response.IDName
	for _, u := range resp.Users {
		users = append(users, response.IDName{
			ID:   u.Id,
			Name: u.Name,
		})
	}
	return &response.Users{
		Users: users,
	}, nil
}

func ChangeUserStatus(ctx *gin.Context, userID, orgID string, status bool) error {
	_, err := iam.ChangeUserStatus(ctx.Request.Context(), &iam_service.ChangeUserStatusReq{
		UserId: userID,
		OrgId:  orgID,
		Status: status,
	})
	return err
}

func ChangeUserPassword(ctx *gin.Context, userID string, req *request.UserPassword) error {
	oldPassword, err := decryptCipherRSA(ctx.Request.Context(), req.OldCipher, req.KeyID, challengeValidateOnly)
	if err != nil {
		return fmt.Errorf("decrypt old password err: %v", err)
	}
	newPassword, err := decryptCipherRSA(ctx.Request.Context(), req.NewCipher, req.KeyID, challengeConsume)
	if err != nil {
		return fmt.Errorf("decrypt new password err: %v", err)
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	_, err = iam.UpdateUserPassword(ctx.Request.Context(), &iam_service.UpdateUserPasswordReq{
		UserId:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	return err
}

func AdminChangeUserPassword(ctx *gin.Context, userID string, req *request.UserPasswordByAdmin) error {
	password, err := decryptCipherRSA(ctx.Request.Context(), req.Cipher, req.KeyID, challengeConsume)
	if err != nil {
		return fmt.Errorf("decrypt password err: %v", err)
	}
	_, err = iam.ResetUserPassword(ctx.Request.Context(), &iam_service.ResetUserPasswordReq{
		UserId:   userID,
		Password: password,
	})
	return err
}

func GetOrgUserNotSelect(ctx *gin.Context, orgID, name string) (*response.UserSelect, error) {
	users, err := iam.GetUserSelectNotInOrg(ctx.Request.Context(), &iam_service.GetUserSelectNotInOrgReq{
		OrgId:    orgID,
		UserName: name,
	})
	if err != nil {
		return nil, err
	}
	return &response.UserSelect{Select: toIDNames(users.Selects)}, nil
}

func GetRoleSelect(ctx *gin.Context, orgID string) (*response.RoleSelect, error) {
	roles, err := iam.GetRoleSelect(ctx.Request.Context(), &iam_service.GetRoleSelectReq{
		OrgId: orgID,
	})
	if err != nil {
		return nil, err
	}
	selects := toRoleIDNames(ctx, roles.Roles)
	// also include global roles
	globalRoles, err := iam.GetGlobalRoleSelect(ctx.Request.Context(), &iam_service.GetGlobalRoleSelectReq{})
	if err != nil {
		return nil, err
	}
	selects = append(selects, toRoleIDNames(ctx, globalRoles.Roles)...)
	return &response.RoleSelect{Select: selects}, nil
}

func AddOrgUser(ctx *gin.Context, orgID, userID, roleID string) error {
	_, err := iam.AddOrgUser(ctx.Request.Context(), &iam_service.AddOrgUserReq{
		OrgId:  orgID,
		UserId: userID,
		RoleId: roleID,
	})
	return err
}

func RemoveOrgUser(ctx *gin.Context, orgID, userID string) error {
	_, err := iam.RemoveOrgUser(ctx.Request.Context(), &iam_service.RemoveOrgUserReq{
		OrgId:  orgID,
		UserId: userID,
	})
	return err
}

func UpdateUserAvatar(ctx *gin.Context, userID, key string) error {
	_, err := iam.UpdateUserAvatar(ctx.Request.Context(), &iam_service.UpdateUserAvatarReq{
		UserId:     userID,
		AvatarPath: key,
	})
	return err
}

func CreateUserByFile(ctx *gin.Context, creatorID, orgID string) (*response.UserBatchImportResult, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_user_batch_import_file", fmt.Sprintf("get file err: %v", err))
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_user_batch_import_file", fmt.Sprintf("open file err: %v", err))
	}
	defer func() { _ = file.Close() }()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_user_batch_import_file", fmt.Sprintf("read file err: %v", err))
	}

	users, err := parseUserExcel(fileBytes)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_user_batch_import_file", fmt.Sprintf("parse excel err: %v", err))
	}

	// BFF层校验（收集错误，不中断）
	// validUsers: 保存有效用户及其行号
	type userWithRow struct {
		user *iam_service.CreateUsersInfo
		row  int // 行号
	}
	var validUsers []userWithRow
	var skippedRows int // 跳过的空行数
	result := &response.UserBatchImportResult{}
	for i, user := range users {
		// 跳过空行（所有字段都为空的行）
		if user.UserName == "" && user.Password == "" && user.Phone == "" && user.Email == "" && user.RoleName == "" {
			skippedRows++
			continue
		}

		row := i + 2 // Excel行号（第1行是表头）

		if err := validateUsername(user.UserName); err != nil {
			result.Errors = append(result.Errors, response.UserBatchImportError{
				Row:      row,
				Username: user.UserName,
				Reason:   err.Error(),
			})
			continue
		}
		if err := validatePassword(user.Password); err != nil {
			result.Errors = append(result.Errors, response.UserBatchImportError{
				Row:      row,
				Username: user.UserName,
				Reason:   err.Error(),
			})
			continue
		}
		if config.Cfg().CustomInfo.UserPhoneRequired != 0 && user.Phone == "" {
			result.Errors = append(result.Errors, response.UserBatchImportError{
				Row:      row,
				Username: user.UserName,
				Reason:   "电话号码不能为空",
			})
			continue
		}
		if user.Phone != "" {
			if err := validatePhone(user.Phone); err != nil {
				result.Errors = append(result.Errors, response.UserBatchImportError{
					Row:      row,
					Username: user.UserName,
					Reason:   err.Error(),
				})
				continue
			}
		}
		if user.Email != "" {
			if err := validateEmail(user.Email); err != nil {
				result.Errors = append(result.Errors, response.UserBatchImportError{
					Row:      row,
					Username: user.UserName,
					Reason:   err.Error(),
				})
				continue
			}
		}

		validUsers = append(validUsers, userWithRow{
			user: user,
			row:  row, // 保存行号
		})
	}
	result.Total = len(users) - skippedRows
	if result.Total > MaxBatchCreateUsersLimit {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_user_batch_import_file", fmt.Sprintf("批量创建用户条数不能超过%d条", MaxBatchCreateUsersLimit))
	}

	if len(validUsers) > 0 {
		// 构建请求，只传递有效用户
		var validUsersInfo []*iam_service.CreateUsersInfo
		for _, v := range validUsers {
			validUsersInfo = append(validUsersInfo, v.user)
		}

		resp, err := iam.CreateUsers(ctx.Request.Context(), &iam_service.CreateUsersReq{
			CreatorId: creatorID,
			OrgId:     orgID,
			Users:     validUsersInfo,
		})
		if err != nil {
			return nil, err
		}

		result.Success = int(resp.Success)

		// 合并IAM层的错误
		// IAM返回的index是在validUsers数组中的索引，需要映射回原始行号
		for _, e := range resp.Errors {
			iamIndex := int(e.Index)
			// iamIndex是validUsers数组的索引
			if iamIndex >= 0 && iamIndex < len(validUsers) {
				result.Errors = append(result.Errors, response.UserBatchImportError{
					Row:      validUsers[iamIndex].row,
					Username: validUsers[iamIndex].user.UserName,
					Reason:   e.Reason,
				})
			}
		}
	}

	// 按行号排序
	sort.Slice(result.Errors, func(i, j int) bool {
		return result.Errors[i].Row < result.Errors[j].Row
	})

	result.Failed = len(result.Errors)

	return result, nil
}

func parseUserExcel(fileData []byte) ([]*iam_service.CreateUsersInfo, error) {
	wb, err := util.OpenWorkbookFromBytes(fileData)
	if err != nil {
		return nil, err
	}
	defer func() { _ = wb.Close() }()

	sheets, err := wb.GetSheets()
	if err != nil {
		return nil, fmt.Errorf("excel has no sheets")
	}
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel has no sheets")
	}
	rows, err := wb.GetRows("")
	if err != nil {
		return nil, fmt.Errorf("invalid excel data")
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("excel has no data rows")
	}
	headerRow := rows[0]
	headerSet := make(map[string]bool)
	for _, h := range headerRow {
		headerSet[h] = true
	}
	for _, required := range requiredUserExcelHeaders {
		if !headerSet[required] {
			return nil, fmt.Errorf("excel header invalid: missing %s", required)
		}
	}

	records, err := wb.ReadWithHeaderMapping(util.ReadWithHeaderMappingOptions{
		Sheet:     "",
		HeaderRow: 0,
		HeaderMapping: map[string]string{
			ExcelHeaderUserName: "userName",
			ExcelHeaderPassword: "password",
			ExcelHeaderPhone:    "phone",
			ExcelHeaderRole:     "roleName",
			ExcelHeaderEmail:    "email",
		},
	})
	if err != nil {
		return nil, err
	}

	var users []*iam_service.CreateUsersInfo
	for _, record := range records {
		// 保留所有行（包括userName为空的行），保持索引与Excel行号对应
		users = append(users, &iam_service.CreateUsersInfo{
			UserName: record["userName"],
			Password: record["password"],
			Phone:    record["phone"],
			Email:    record["email"],
			RoleName: record["roleName"],
		})
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no valid user data")
	}
	return users, nil
}

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9\x{4e00}-\x{9fa5}_().]+$`)
	phoneRegex    = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func validateUsername(username string) error {
	if len(username) < 2 || len(username) > 20 {
		return fmt.Errorf("用户名长度需为2-20个字符")
	}
	if username[0] == '_' {
		return fmt.Errorf("用户名不能以下划线开头")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("用户名只能包含中英文、数字、下划线、括号")
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("请输入密码")
	}
	if len(password) < 8 || len(password) > 20 {
		return fmt.Errorf("密码长度需为8-20个字符")
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasLetter || !hasNumber || !hasSpecial {
		return fmt.Errorf("密码需包含字母、数字、特殊字符")
	}
	return nil
}

func validatePhone(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("电话号码格式不正确")
	}
	return nil
}

func validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("邮箱格式不正确")
	}
	return nil
}

// --- internal ---

func toUserInfo(ctx *gin.Context, user *iam_service.UserInfo) *response.UserInfo {
	ret := &response.UserInfo{
		UserID:    user.UserId,
		Username:  user.UserName,
		Phone:     user.Phone,
		Email:     user.Email,
		CreatedAt: util.Time2Str(user.CreatedAt),
		Creator:   toIDName(user.Creator),
		Status:    user.Status,
		Language:  getLanguageByCode(user.Language),
		Avatar:    cacheUserAvatar(ctx, user.AvatarPath),
	}
	for _, userOrg := range user.Orgs {
		ret.Orgs = append(ret.Orgs, toOrgRole(ctx, userOrg))
	}
	return ret
}

func toOrgRole(ctx *gin.Context, userOrg *iam_service.UserOrg) response.OrgRole {
	return response.OrgRole{
		Org:   toOrgIDName(ctx, userOrg.Org),
		Roles: toRoleIDNames(ctx, userOrg.Roles),
	}
}

// rsaCipherPayload RSA加密传输的载荷结构
// 前端将 {password, challenge} 打包为JSON后整体RSA加密
type rsaCipherPayload struct {
	Password  string `json:"password"`  // 明文密码
	Challenge string `json:"challenge"` // 服务端下发的Challenge
}

const (
	// challengeConsume 消费Challenge，校验后立即删除（GET+DEL），用于单cipher场景或双cipher的最后一次解密
	challengeConsume = true
	// challengeValidateOnly 仅校验Challenge是否存在，不消费，用于双cipher场景中非最后一次解密
	challengeValidateOnly = false
)

// decryptCipherRSA 使用RSA解密cipher并校验Challenge
// cipher = RSA-OAEP-SHA256(Base64(JSON({password, challenge})), publicKey)
//
// 流程：
//  1. RSA私钥解密cipher，得到JSON明文
//  2. 解析JSON提取password和challenge
//  3. 校验challenge：consume=true时一次性消费（GET+DEL），consume=false时仅校验存在性
//
// 返回解密后的明文密码
func decryptCipherRSA(ctx context.Context, cipher string, keyID string, consume bool) (string, error) {
	// 1. RSA解密
	plaintextBytes, err := rsautil.GetManager().Decrypt(keyID, cipher)
	if err != nil {
		return "", fmt.Errorf("RSA decrypt failed: %w", err)
	}

	// 2. 解析JSON载荷
	var payload rsaCipherPayload
	if err := json.Unmarshal(plaintextBytes, &payload); err != nil {
		return "", fmt.Errorf("invalid cipher payload: %w", err)
	}

	// 3. 校验Challenge
	challengeManager := bff_rsautil.GetChallengeManager()
	if consume {
		// 消费Challenge（Redis原子GET+DEL，一次性），防重放攻击
		ok, err := challengeManager.ValidateAndConsume(ctx, payload.Challenge)
		if err != nil {
			return "", fmt.Errorf("challenge validation failed: %w", err)
		}
		if !ok {
			return "", fmt.Errorf("challenge invalid or already used, possible replay attack")
		}
	} else {
		// 仅校验Challenge是否存在，不消费（用于同一个challenge解密多个cipher的场景）
		ok, err := challengeManager.Validate(ctx, payload.Challenge)
		if err != nil {
			return "", fmt.Errorf("challenge validation failed: %w", err)
		}
		if !ok {
			return "", fmt.Errorf("challenge invalid or expired, possible replay attack")
		}
	}

	return payload.Password, nil
}
