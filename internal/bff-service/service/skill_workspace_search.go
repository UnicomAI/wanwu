package service

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	path_util "github.com/UnicomAI/wanwu/pkg/path-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// searcher 封装单次搜索的配置和状态。
type searcher struct {
	workspaceDir   string
	keyword        string
	caseSensitive  bool
	wholeWord      bool
	useRegex       bool
	compiledRegex  *regexp.Regexp
	includePattern string
	excludePattern string
}

// newSearcher 根据请求参数构造工作区搜索器。
func newSearcher(workspaceDir string, req request.SearchSkillWorkspaceReq) (*searcher, error) {
	s := &searcher{
		workspaceDir:   workspaceDir,
		keyword:        req.Keyword,
		caseSensitive:  req.CaseSensitive,
		wholeWord:      req.WholeWord,
		useRegex:       req.UseRegex,
		compiledRegex:  req.CompiledRegex,
		includePattern: req.IncludePattern,
		excludePattern: req.ExcludePattern,
	}
	// 保险：若调用方绕过 Check() 直接构造请求，仍可安全编译
	if req.UseRegex && s.compiledRegex == nil {
		flags := ""
		if !req.CaseSensitive {
			flags = "(?i)"
		}
		compiled, err := regexp.Compile(flags + req.Keyword)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %v", err)
		}
		s.compiledRegex = compiled
	}
	return s, nil
}

// shouldVisit 判断路径是否需要进入搜索。
func (s *searcher) shouldVisit(relPath string, info fs.FileInfo) bool {
	// 跳过隐藏目录/文件
	for _, part := range strings.Split(filepath.ToSlash(relPath), "/") {
		if part != "." && strings.HasPrefix(part, ".") {
			return false
		}
	}
	// 跳过 node_modules
	if info.IsDir() && info.Name() == "node_modules" {
		return false
	}
	return true
}

// matchLine 判断单行内容是否命中搜索条件。
func (s *searcher) matchLine(line string) bool {
	if s.useRegex {
		return s.compiledRegex.MatchString(line)
	}
	searchLine := line
	searchKeyword := s.keyword
	if !s.caseSensitive {
		searchLine = strings.ToLower(line)
		searchKeyword = strings.ToLower(s.keyword)
	}
	if s.wholeWord {
		for _, word := range strings.FieldsFunc(searchLine, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
		}) {
			if word == searchKeyword {
				return true
			}
		}
		return false
	}
	return strings.Contains(searchLine, searchKeyword)
}

// scanFile 扫描单个文本文件并返回匹配行。
func (s *searcher) scanFile(path, relPath string) ([]*response.SearchResult, error) {
	if binary, err := util.IsLikelyBinaryFile(path); err != nil || binary {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var results []*response.SearchResult
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), maxSearchLineBytes)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if s.matchLine(line) {
			content := line
			if len(content) > maxSearchResultContentLength {
				content = util.TruncateUTF8(content, maxSearchResultContentLength)
			}
			results = append(results, &response.SearchResult{
				Path:    relPath,
				Line:    lineNum,
				Content: content,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// run 遍历工作区并执行搜索。
func (s *searcher) run() (*response.SkillWorkspaceSearchResp, error) {
	results := make([]*response.SearchResult, 0)
	truncated := false

	err := filepath.WalkDir(s.workspaceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // 跳过不可访问的路径
		}
		relPath, err := filepath.Rel(s.workspaceDir, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)
		// 兜底解码存量 GBK 文件名：磁盘字节当 UTF-8 解码会产生 U+FFFD，
		// 这里转为 UTF-8，使搜索结果 path 与 files 接口 path 编码一致，前端可复用。
		// 注意磁盘访问用原始 path（os.Open/读内容），仅结果 path 用解码后的 relPath。
		relPath = util.DecodeGBKToUTF8(relPath)

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}

		if !s.shouldVisit(relPath, info) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if s.excludePattern != "" && util.MatchGlobPatterns(relPath, s.excludePattern) {
			return nil
		}
		if s.includePattern != "" && !util.MatchGlobPatterns(relPath, s.includePattern) {
			return nil
		}
		if info.Size() > maxFileSize {
			return nil
		}

		fileResults, err := s.scanFile(path, relPath)
		if err != nil {
			return nil // 跳过不可读文件
		}
		for _, r := range fileResults {
			if len(results) >= maxSearchResults {
				truncated = true
				return filepath.SkipAll
			}
			results = append(results, r)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk error: %w", err)
	}

	return &response.SkillWorkspaceSearchResp{
		Results:   results,
		Total:     len(results),
		Truncated: truncated,
	}, nil
}

// SearchInWorkspace 在 Skill 工作区中搜索关键词。
func SearchInWorkspace(ctx *gin.Context, userId, orgId string, req request.SearchSkillWorkspaceReq) (*response.SkillWorkspaceSearchResp, error) {
	ws, err := resolveSkillWorkspace(req.CustomSkillID)
	if err != nil {
		return nil, err
	}
	if err := path_util.EnsureNoSymlinkInPath(ws.workspaceDir, ws.workspaceDir, true); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	if _, err := os.Stat(ws.workspaceDir); os.IsNotExist(err) {
		return &response.SkillWorkspaceSearchResp{Results: []*response.SearchResult{}, Total: 0}, nil
	}

	s, err := newSearcher(ws.workspaceDir, req)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, err.Error())
	}
	resp, err := s.run()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_workspace_search_failed")
	}
	return resp, nil
}
