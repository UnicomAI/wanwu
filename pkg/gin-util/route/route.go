package route

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var paths sync.Map

type PermLevel uint32

const (
	PermNone       PermLevel = 0 // No verification required
	PermNeedEnable PermLevel = 1 // Requires user to enable
	PermNeedCheck  PermLevel = 2 // Requires user permissions
)

type Route struct {
	Tag  string
	Name string
	Subs []Route
}

type TagName struct {
	Tag  string
	Name string
}

type Paths []*_path

func (paths Paths) Get(absPath, method string) *_path {
	for _, path := range paths {
		if path.absPath == absPath && path.method == method {
			return path
		}
	}
	return nil
}

// --- API ---

func LoadOrStore(absPath, method, desc string, authType PermLevel, tag TagName, handler gin.HandlerFunc, middlewares ...gin.HandlerFunc) (bool, error) {
	key := pathKey(absPath, method)
	actual, loaded := paths.LoadOrStore(key, &_path{
		absPath:  absPath,
		method:   method,
		desc:     desc,
		authType: authType,
		tags:     []TagName{tag},

		handler:     handler,
		middlewares: middlewares,
	})
	if loaded {
		path := actual.(*_path)
		// desc
		if path.desc != desc {
			return loaded, fmt.Errorf("%v desc %v conflict with %v", key, path.desc, desc)
		}
		// authType
		if path.authType != authType {
			return loaded, fmt.Errorf("%v authType %v conflict with %v", key, path.authType, authType)
		}
		// handler
		if reflect.ValueOf(path.handler).Pointer() != reflect.ValueOf(handler).Pointer() {
			return loaded, fmt.Errorf("%v handler %v conflict with %v", key,
				runtime.FuncForPC(reflect.ValueOf(path.handler).Pointer()).Name(),
				runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
		}
		// middlewares
		if len(path.middlewares) != len(middlewares) {
			return loaded, fmt.Errorf("%v middlewares len conflict", key)
		}
		for _, middleware := range middlewares {
			var exist bool
			for _, m := range path.middlewares {
				if reflect.ValueOf(m).Pointer() == reflect.ValueOf(middleware).Pointer() {
					exist = true
					break
				}
				parts1 := strings.Split(runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name(), ".")
				parts2 := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")
				if len(parts1) == len(parts2) && len(parts1) >= 2 && parts1[len(parts1)-2] == parts2[len(parts2)-2] {
					// Middleware may be a closure func object
					exist = true
					break
				}
			}
			if !exist {
				return loaded, fmt.Errorf("%v middleware %v conflict, not exist", key,
					runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name())
			}
		}
		// tag
		for _, r := range path.tags {
			if r == tag {
				return loaded, fmt.Errorf("%v tag %v already exist", key, tag)
			}
		}
		path.tags = append(path.tags, tag)
	}
	return loaded, nil
}

func GetTags(absPath, method string) ([]string, bool) {
	v, loaded := paths.Load(pathKey(absPath, method))
	if !loaded {
		return nil, false
	}
	p, ok := v.(*_path)
	if !ok {
		return nil, false
	}
	var ret []string
	for _, tag := range p.tags {
		ret = append(ret, tag.Tag)
	}
	return ret, true
}

type _path struct {
	absPath  string
	method   string
	desc     string
	authType PermLevel
	tags     []TagName

	handler     gin.HandlerFunc
	middlewares []gin.HandlerFunc
}

func pathKey(absPath, method string) string {
	return fmt.Sprintf("[%s]%s", method, absPath)
}
