package service

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

const (
	docCenterLocalDir        = "configs/microservice/bff-service/static/manual/"
	docCenterStaticAPIPrefix = "../../../user/api/v1/static/manual" // ../../.. is used to offset the front-end fixed prefix aibase/docCenter/pages
	docCenterSnippetLen      = 200                                  // Cut text length

	// Common naming for documents:
	// fileName:    e.g. StartNode.md
	// filePath:    configs/microservice/bff-service/static/manual + relFilePath e.g. configs/microservice/bff-service/static/manual/workflow/StartNode.md
	// relFilePath: relative path to the file in configs/microservice/bff-service/static/manual e.g. workflow/StartNode.md
)

var (
	mdImageRegex          = regexp.MustCompile(`!\[.*?\]\((.*?)\)`)        // Match ![](xxxxx) image quotes from markdown text
	mdParenthesisRefRegex = regexp.MustCompile(`\((.*?)\)`)                // Match (xxxxx) from markdown quote
	mdLinkRegex           = regexp.MustCompile(`[^!]\[.*?\]\((.*?\.md)\)`) // Match jump links from markdown [](xxxxx)
	mdBracketRegex        = regexp.MustCompile(`\[(.*?)\]`)                // Match text in [] from markdown

	_docCenter *docCenter
)

type docCenter struct {
	menus    []*response.DocMenu
	contents map[string]string // refFilePath -> content
	searcher *riot.Engine
}

type mdInfo struct {
	relFilePath string
	content     string
}

func InitDocCenter() error {
	if _docCenter != nil {
		return errors.New("already init")
	}

	// 0. Read all md files in docCenterLocalDir
	var mdInfos []mdInfo
	if err := filepath.Walk(docCenterLocalDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it is not a directory and if it is a markdown file
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".md") {
			// Read file contents
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read %v err: %v", filePath, err)
			}
			// Process the xxxxx in the image reference ![](xxxxx) and link reference [](xxxxx) in the markdown text as an address accessible to the front end
			relFilePath := strings.TrimPrefix(filePath, docCenterLocalDir)
			convertByte := convertMarkdown(docCenterStaticAPIPrefix, relFilePath, string(content))
			mdInfos = append(mdInfos, mdInfo{
				content:     convertByte,
				relFilePath: relFilePath,
			})
		}
		return nil
	}); err != nil {
		return err
	}

	// 1. Build a search engine
	searcher, err := newDocMenuSearcher(mdInfos)
	if err != nil {
		return fmt.Errorf("init search engin err: %v", err)
	}

	// 2. Construct menus and contents
	var menus []*response.DocMenu
	contents := make(map[string]string)
	for _, mdInfo := range mdInfos {
		contents[mdInfo.relFilePath] = mdInfo.content
		addDocMenusMdFile(&menus, mdInfo.relFilePath, mdInfo)
	}

	// 3. Refresh the index
	// 3.1 menu sorting
	sortDocMenus(&menus)
	// 3.2 Regenerate menu index
	for i, menu := range menus {
		refreshDocMenuIndex(menu, fmt.Sprintf("doc%d", i+1))
	}

	_docCenter = &docCenter{
		menus:    menus,
		contents: contents,
		searcher: searcher,
	}
	return nil
}

func GetDocCenterMenu(ctx *gin.Context) []*response.DocMenu {
	return _docCenter.menus
}

func SearchDocCenter(ctx *gin.Context, content string) ([]response.DocSearchResp, error) {
	results := _docCenter.searcher.SearchDoc(types.SearchReq{Text: content})
	var searchResps []response.DocSearchResp
	for _, doc := range results.Docs {
		title := strings.TrimSuffix(filepath.Base(doc.DocId), filepath.Ext(filepath.Base(doc.DocId)))
		snippet, err := util.Md2html([]byte(getMarkdownSnippet(doc.Content, content, docCenterSnippetLen)))
		if err != nil {
			log.Errorf("doc center %v md2html error", doc.DocId)
			continue // Skip the current doc without processing it
		}
		searchUrl, err := url.JoinPath(config.Cfg().Server.WebBaseUrl, config.Cfg().DocCenter.FrontendPrefix, url.PathEscape(doc.DocId))
		if err != nil {
			log.Errorf("doc center %v to search url err: %v", doc.DocId, err)
			continue
		}
		searchResp := response.DocSearchResp{
			Title: title,
			ContentList: []response.DocSearchContent{
				{
					Title:   title,
					Content: snippet,
					Url:     searchUrl,
				},
			},
		}
		searchResps = append(searchResps, searchResp)
	}
	return searchResps, nil
}

func GetDocCenterMarkdown(ctx *gin.Context, relFilePath string) (string, error) {
	relFilePath, err := url.QueryUnescape(relFilePath)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_doc_center_file_unescape", err.Error())
	}
	// check fileName
	if !strings.HasSuffix(relFilePath, ".md") {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_doc_center_file_md", relFilePath)
	}
	return _docCenter.contents[relFilePath], nil
}

// --- doc-center convert raw markdown ---

// Process the xxxxx in the image reference ![](xxxxx) and link reference [](xxxxx) in the markdown text as an address accessible to the front end
//
//nolint:staticcheck
func convertMarkdown(apiPrefix, refFilePath, mdContent string) string {
	convertHttp := mdLinkRegex.ReplaceAllStringFunc(mdContent, func(mdLabel string) string {
		for _, httpRelPaths := range mdParenthesisRefRegex.FindAllStringSubmatch(mdLabel, -1) {
			if len(httpRelPaths) <= 1 {
				return mdLabel
			}
			txt := mdBracketRegex.FindString(mdLabel)
			return txt + "(" + url.PathEscape(path.Join(refFilePath, "../", httpRelPaths[1])) + ")"
		}
		return mdLabel
	})
	return mdImageRegex.ReplaceAllStringFunc(convertHttp, func(imageLabel string) string {
		// imageLabel is the matched image format, such as ![](../assets/append.png)
		for _, imageRelPaths := range mdParenthesisRefRegex.FindAllStringSubmatch(imageLabel, -1) {
			// Continue matching from imageLabel, for example, imageRelPaths[0] is ![](../assets/append.png), imageRelPaths[1] is ../assets/append.png
			if len(imageRelPaths) <= 1 {
				return imageLabel
			}
			// Regenerate the image reference and process ../assets/append.png as user/api/v1/static/manual/assets/append.png
			// For example, refFilePath is workflow/StartNode.md
			// 1. "workflow/StartNode.md" + "../" + "../assets/append.png" => assets/append.png
			// 2. "../../../user/api/v1/static/manual" + "assets/append.png" => ../../../user/api/v1/static/manual/assets/append.png
			// 3. Escape non-numeric letters in the path, and then convert %2F back to /
			return "![](" + strings.ReplaceAll(url.PathEscape(path.Join(apiPrefix, path.Join(refFilePath, "../", imageRelPaths[1]))), "%2F", "/") + ")"
		}
		return imageLabel
	})
}

// --- doc-center search engine ---

func newDocMenuSearcher(mdInfos []mdInfo) (*riot.Engine, error) {
	engine := &riot.Engine{}
	engine.Init(types.EngineOpts{})
	// Create index
	for _, mdInfo := range mdInfos {
		engine.Index(mdInfo.relFilePath, types.DocData{
			Content: string(mdInfo.content),
		})
	}
	// refresh index
	engine.Flush()
	return engine, nil
}

func getMarkdownSnippet(content, keyword string, snippetLen int) string {
	//String is a read-only byte slice encoded with utf8. Therefore, the length obtained by using the len function is not the number of characters, but the number of bytes.
	//rune is an alias of int32, which represents the Unicode encoding of the character and is stored in 4 bytes. Converting string to rune means that any character uses 4 bytes to store its unicode value.
	//In this way, the unicode value returned each time it is traversed is no longer bytes.
	runes := []rune(content)
	keyRunes := []rune(keyword)
	index := strings.Index(content, keyword)
	if index == -1 {
		if len(runes) < snippetLen {
			return string(runes)
		} else {
			return string(runes[0:snippetLen])
		}
	}
	runeIndex := len([]rune(content[:index]))
	start := runeIndex - snippetLen
	if start < 0 {
		start = 0
	}
	end := runeIndex + len(keyRunes) + snippetLen
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// --- doc-center add markdown file info to menus ---

func addDocMenusMdFile(menus *[]*response.DocMenu, rest string, mdInfo mdInfo) {
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 {
		return
	}
	var menu *response.DocMenu
	for _, curr := range *menus {
		if curr.Name == parts[0] {
			menu = curr
			break
		}
	}
	if menu == nil {
		menu = &response.DocMenu{
			Name: parts[0],
		}
		if len(parts) == 1 { // Not a directory, it is an md file
			menu.Name = strings.TrimSuffix(menu.Name, ".md")
			menu.PathRaw = mdInfo.relFilePath
			menu.Path = url.PathEscape(mdInfo.relFilePath) // The front end requires path escaping
			menu.SetContent(mdInfo.content)
		}
		*menus = append(*menus, menu)
	}
	if len(parts) > 1 {
		addDocMenusMdFile(&(menu.Children), parts[1], mdInfo)
	}
}

// --- doc-center sort menus ---

func sortDocMenus(menus *[]*response.DocMenu) {
	sort.Slice(*menus, func(i, j int) bool {
		return orderDocNum((*menus)[i].Name, (*menus)[j].Name)
	})
	for _, menu := range *menus {
		if len(menu.Children) > 0 {
			sortDocMenus(&menu.Children)
		}
	}
}

// Implement natural ordering (number first)
func orderDocNum(s1, s2 string) bool {
	numParts1, isNum1 := extractDocNum(s1)
	numParts2, isNum2 := extractDocNum(s2)
	if isNum1 && isNum2 {
		return numParts1 < numParts2
	} else if isNum1 {
		// If one is a number and one is a non-number, the numeric part comes first
		return true
	} else if isNum2 {
		return false
	}
	return s1 < s2
}

// extractNum splits the string into numeric and non-numeric parts
func extractDocNum(s string) (int, bool) {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			break
		}
	}
	num, err := strconv.Atoi(result.String())
	if err != nil {
		return 0, false
	}
	return num, true
}

// --- doc-center refresh index ---

func refreshDocMenuIndex(menu *response.DocMenu, index string) {
	menu.Index = index
	for i, child := range menu.Children {
		refreshDocMenuIndex(child, fmt.Sprintf("%s-%d", index, i+1))
	}
}
