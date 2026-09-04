package ahocorasick

import (
	"fmt"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	dictLRUDura       = time.Hour
	dictLRUTickerDura = 30 * time.Minute
)

var (
	_ac *acMgr
)

type acDict struct {
	DictID   string
	Version  string
	Reply    string
	LastUsed time.Time
	Matcher  *Matcher
}

type acMgr struct {
	mu       sync.RWMutex
	dicts    map[string]*acDict
	stopChan chan struct{}
}

type MatchResult struct {
	DictCfg   DictConfig
	WordIndex int
	Word      string
	Reply     string
}

type DictConfig struct {
	DictID  string
	Version string
}

type DictStatus struct {
	DictCfg DictConfig
	Status  bool // 词表是否存在
}

// ------------ api -----------
func BuildDict(dict DictConfig, reply string, words []string) error {
	if _ac == nil {
		return fmt.Errorf("aho-corasick not init")
	}
	return _ac.buildDict(dict, reply, words)
}

func CheckDictStatus(dicts []DictConfig) ([]DictStatus, error) {
	if _ac == nil {
		return nil, fmt.Errorf("aho-corasick not init")
	}
	return _ac.checkDictStatus(dicts)
}

func ContentMatch(content string, dicts []DictConfig, returnFirstMatch bool) ([]MatchResult, error) {
	if _ac == nil {
		return nil, fmt.Errorf("aho-corasick not init")
	}
	return _ac.contentMatch(content, dicts, returnFirstMatch)
}

func ContentContain(content string, dicts []DictConfig) (*DictConfig, error) {
	if _ac == nil {
		return nil, fmt.Errorf("aho-corasick not init")
	}
	return _ac.contentContain(content, dicts)
}

// ------------ init -----------
func Init(useLRU bool) error {
	if _ac != nil {
		return fmt.Errorf("aho-corasick already init")
	}
	_ac = &acMgr{
		dicts:    make(map[string]*acDict),
		stopChan: make(chan struct{}),
	}
	if useLRU {
		_ac.startCleanup()
	}
	return nil
}

func Stop() {
	if _ac != nil {
		_ac.stop()
	}
	_ac = nil
}

// ------------ internal -----------
func (a *acMgr) startCleanup() {
	ticker := time.NewTicker(dictLRUTickerDura)
	go func() {
		defer util.PrintPanicStack()
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.mu.Lock()
				threshold := time.Now().Add(-1 * dictLRUDura)
				for k, v := range a.dicts {
					log.Debugf("aho-corasick dict [%v]", k)
					if v.LastUsed.Before(threshold) {
						log.Debugf("aho-corasick dict delete [%v] lastUsed[%v]", k, v.LastUsed)
						delete(a.dicts, k)
					}
				}
				a.mu.Unlock()
			case <-a.stopChan:
				return
			}
		}
	}()
}

func (a *acMgr) stop() {
	close(a.stopChan)
}

func (a *acMgr) buildDict(dict DictConfig, reply string, words []string) error {
	if dict.DictID == "" || dict.Version == "" {
		return fmt.Errorf("aho-corasick build dict,config can not be empty")
	}
	dictKey := getDictKey(dict.DictID, dict.Version)
	matcher := NewStringMatcher(words)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dicts[dictKey] = &acDict{
		DictID:   dict.DictID,
		Version:  dict.Version,
		Matcher:  matcher,
		Reply:    reply,
		LastUsed: time.Now(),
	}
	return nil
}

func (a *acMgr) checkDictStatus(dicts []DictConfig) ([]DictStatus, error) {
	if len(dicts) == 0 {
		return nil, nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]DictStatus, 0, len(dicts))
	for _, cfg := range dicts {
		if cfg.Version == "" || cfg.DictID == "" {
			return nil, fmt.Errorf("aho-corasick check dict status,config id or version can not be empty")
		}
		dictKey := getDictKey(cfg.DictID, cfg.Version)
		_, exists := a.dicts[dictKey]
		if exists {
			a.dicts[dictKey].LastUsed = time.Now()
		}
		result = append(result, DictStatus{
			DictCfg: cfg,
			Status:  exists,
		})
	}
	return result, nil
}

func (a *acMgr) contentMatch(content string, dicts []DictConfig, returnFirstMatch bool) ([]MatchResult, error) {
	if len(content) == 0 || len(dicts) == 0 {
		return nil, nil
	}
	// 1. 加读锁，仅保护字典读取阶段
	a.mu.RLock()
	dictsCopy := make(map[string]*acDict) // 复制字典，避免后续操作依赖锁
	for _, cfg := range dicts {
		dictKey := getDictKey(cfg.DictID, cfg.Version)
		if dict, exists := a.dicts[dictKey]; exists && dict != nil {
			a.dicts[dictKey].LastUsed = time.Now()
			dictsCopy[dictKey] = dict
		}
	}
	a.mu.RUnlock() // 字典读取完毕，立即释放锁
	if len(dictsCopy) != len(dicts) {
		log.Warnf("aho-corasick content match dictscopy len[%v] dicts len[%v] missmatch", len(dictsCopy), len(dicts))
	}
	// 2. 无锁状态下进行匹配
	results := make([]MatchResult, 0)
	contentBytes := []byte(content)
	for _, cfg := range dicts {
		dictKey := getDictKey(cfg.DictID, cfg.Version)
		if dict, exists := dictsCopy[dictKey]; exists {
			matches := dict.Matcher.MatchThreadSafeWithPos(contentBytes)
			if len(matches) > 0 {
				for _, match := range matches {
					if !isWordBoundaryMatch(contentBytes, match.Start, match.End, match.WordBytes) {
						continue
					}
					matchResult := MatchResult{
						DictCfg: DictConfig{
							DictID:  dict.DictID,
							Version: dict.Version,
						},
						WordIndex: match.Index,
						Word:      dict.Matcher.getOriginalWord(match.Index),
						Reply:     dict.Reply,
					}
					results = append(results, matchResult)
					if returnFirstMatch {
						return results, nil
					}
				}
			}
		}
	}
	return results, nil
}

func (a *acMgr) contentContain(content string, dicts []DictConfig) (*DictConfig, error) {
	if len(content) == 0 || len(dicts) == 0 {
		return nil, nil
	}
	// 1. 加读锁，仅保护字典读取阶段
	a.mu.RLock()
	dictsCopy := make(map[string]*acDict) // 复制字典，避免后续操作依赖锁
	for _, cfg := range dicts {
		dictKey := getDictKey(cfg.DictID, cfg.Version)
		if dict, exists := a.dicts[dictKey]; exists && dict != nil {
			a.dicts[dictKey].LastUsed = time.Now()
			dictsCopy[dictKey] = dict
		}
	}
	a.mu.RUnlock() // 字典读取完毕，立即释放锁
	if len(dictsCopy) != len(dicts) {
		log.Warnf("aho-corasick content match dictscopy len[%v] dicts len[%v] missmatch", len(dictsCopy), len(dicts))
	}
	// 2. 无锁状态匹配字段
	contentBytes := []byte(content)
	for _, cfg := range dictsCopy {
		dictKey := getDictKey(cfg.DictID, cfg.Version)
		if dict, exists := dictsCopy[dictKey]; exists && dict.Matcher.Contains(contentBytes) {
			return &DictConfig{DictID: dict.DictID, Version: dict.Version}, nil
		}
	}
	return nil, nil
}

func (m *Matcher) getOriginalWord(index int) string {
	for _, node := range m.trie {
		if node.output && node.index == index {
			return string(node.b)
		}
	}
	return ""
}

// getDictKey 生成字典的唯一键
func getDictKey(dictID, version string) string {
	return fmt.Sprintf("%s-%s", dictID, version)
}
