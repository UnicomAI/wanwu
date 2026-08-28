package agent_tool

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	minio_service "github.com/UnicomAI/wanwu/internal/agent-service/service/minio-service"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	baseSkillDir = "tmp/skills"
)

type SkillDir struct {
	SkillDir  string //技能所在地址
	RunDir    string //技能运行时目录
	OutputDir string //此次运行技能输出地址
	InputDir  string //此次运行技能输入地址
}

// CreateSkillDir 创建技能运行目录
func CreateSkillDir(runId string, skill *request.SkillToolInfo, uploadFile []string, skillParams *SkillParams) (*SkillDir, error) {
	runDir := baseSkillDir + "/" + runId
	skillDir, err := buildSkillDir(skill)
	if err != nil {
		return nil, err
	}
	//创建inputDir
	var inputDir = runDir + "/inputDir"
	if err := util.MkDir(inputDir); err != nil {
		return nil, err
	}
	if len(uploadFile) > 0 {
		for _, file := range uploadFile {
			err := minio_service.DownloadFileToLocal(context.Background(), file, filepath.Join(inputDir, util.NewRandomFile(file)))
			if err != nil {
				return nil, err
			}
		}
	}
	//下载输入文件
	downloadInputFile(inputDir, skillParams)
	//创建outputDir
	var outputDir = runDir + "/outputDir"
	if err := util.MkDir(outputDir); err != nil {
		return nil, err
	}
	return &SkillDir{
		SkillDir:  skillDir,
		OutputDir: outputDir,
		InputDir:  inputDir,
		RunDir:    runDir,
	}, nil
}

// 构建skill目录
func buildSkillDir(skill *request.SkillToolInfo) (string, error) {
	switch skill.SkillType {
	case request.SkillTypeBuiltIn:
		return skill.ObjectPath, nil
	case request.SkillTypeCustom, request.SkillTypeAcquired:
		return buildPackageSkillDir(skill)
	default:
		return "", fmt.Errorf("unsupported skill type: %s", skill.SkillType)
	}
}

// 构建远程包类型skill目录
func buildPackageSkillDir(skill *request.SkillToolInfo) (string, error) {
	if skill.ObjectPath == "" {
		return "", fmt.Errorf("skill package objectPath is empty, skillType: %s, skillId: %s", skill.SkillType, skill.SkillId)
	}
	// skillType 用于隔离 custom/acquired 两张表中可能相同的 skillId。
	// objectPath 哈希用于避免 skill 重新发布后继续复用旧版本的解压目录。
	skillTempDir := filepath.Join(baseSkillDir, string(skill.SkillType), skill.SkillId, util.MD5([]byte(skill.ObjectPath)))
	exist, err := util.FileExist(skillTempDir)
	if err != nil {
		return "", err
	}
	var skillDir string
	if exist {
		skillDir = skillTempDir
	} else {
		unzipSkill, err := downloadAndUnzipSkill(skillTempDir, skill.ObjectPath)
		if err != nil {
			return "", err
		}
		skillDir = unzipSkill
	}

	fileList, _ := util.DirFileList(skillDir, true, true)
	if len(fileList) > 0 {
		for _, file := range fileList {
			if strings.ToLower(filepath.Base(file)) == "skill.md" {
				return filepath.Dir(file), nil
			}
		}
	}
	return skillDir, nil
}

// 下载并解压skill
func downloadAndUnzipSkill(skillTempDir, skillUrl string) (string, error) {
	//其实这个minio-wanwu的前缀没有实际作用只是为了保持一个http链接格式
	var localFilePath = skillTempDir + "/" + filepath.Base(skillUrl)
	err := minio_service.DownloadFileToLocal(context.Background(), "http://minio-wanwu:9000/"+skillUrl, localFilePath)
	if err != nil {
		return "", err
	}
	return util.UnzipDir(context.Background(), localFilePath, skillTempDir)
}
