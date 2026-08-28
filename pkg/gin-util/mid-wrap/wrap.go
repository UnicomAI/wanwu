package mid_wrap

import (
	"path"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

var _w *wrapper

// --- API ---

func InitWrapper(handlers []gin.HandlerFunc, postHandlers []gin.HandlerFunc) {
	if _w != nil {
		log.Panicf("wrapper already init")
	}
	_w = &wrapper{
		isNavi:          true,
		middlewares:     handlers,
		postMiddlewares: postHandlers,
	}
}

func NewSub(subTag, name string, permLevel route.PermLevel, isNavi, canCheck bool, handlers ...gin.HandlerFunc) {
	if _w == nil {
		log.Panicf("wrapper not init")
	}
	_w.NewSub(subTag, name, permLevel, isNavi, canCheck, handlers...)
}

func Sub(tag string) *wrapper {
	if _w == nil {
		log.Panicf("wrapper not init")
	}
	return _w.sub(tag)
}

func CollectRoutes() []route.Route {
	if _w == nil {
		log.Panicf("wrapper not init")
	}
	var ret []route.Route
	for _, sub := range _w.subs {
		ret = append(ret, sub.collectRoute())
	}
	return ret
}

func CollectPerms() []route.TagName {
	if _w == nil {
		log.Panicf("wrapper not init")
	}
	var ret []route.TagName
	_w.collectPerms(&ret)
	return ret
}

// --- wrapper ---

type wrapper struct {
	tag  string
	name string

	isNavi    bool
	canCheck  bool
	permLevel route.PermLevel

	middlewares     []gin.HandlerFunc
	postMiddlewares []gin.HandlerFunc
	paths           route.Paths

	subs []*wrapper
}

func (w *wrapper) NewSub(subTag, name string, permLevel route.PermLevel, isNavi, canCheck bool, middlewares ...gin.HandlerFunc) {
	if w == nil {
		log.Panicf("wrapper nil")
	}
	// check subTag
	if subTag == "" || strings.Contains(subTag, ".") {
		log.Panicf("wrapper %v subTag %v cannot be empty or contain '.'", w.tag, subTag)
	}
	// check tag 不能与同级的其他sub.tag相同
	var tag string
	if w.tag == "" {
		tag = subTag
	} else {
		tag = w.tag + "." + subTag
	}
	for _, sub := range w.subs {
		if sub.tag == tag {
			log.Panicf("wrapper %v sub %v already exist", w.tag, tag)
		}
	}
	// check isNavi 上级是false，下级必须是false
	if !w.isNavi && isNavi {
		log.Panicf("wrapper %v is not navi but sub %v is navi", w.tag, tag)
	}
	// check permLevel 不能低于上级
	if permLevel < w.permLevel {
		log.Panicf("wrapper %v sub %v permLevel %v cannot less than %v", w.tag, tag, permLevel, w.permLevel)
	}
	// check canCheck 上级是true，下级必须是true
	if w.canCheck && !canCheck {
		log.Panicf("wrapper %v can check but sub %v cannot", w.tag, tag)
	}
	// canCheck是true，permLevel必须大于等于PermNeedCheck
	if canCheck && permLevel < route.PermNeedCheck {
		log.Panicf("wrapper %v sub %v must need perm", w.tag, tag)
	}
	// permLevel小于PermNeedCheck，canCheck必须是false
	if permLevel < route.PermNeedCheck && canCheck {
		log.Panicf("wrapper %v sub %v must cannot check", w.tag, tag)
	}
	w.subs = append(w.subs, &wrapper{
		tag:             tag,
		name:            name,
		isNavi:          isNavi,
		canCheck:        canCheck,
		permLevel:       permLevel,
		middlewares:     append(w.middlewares, middlewares...),
		postMiddlewares: w.postMiddlewares,
	})
}

func (w *wrapper) Reg(rg *gin.RouterGroup, relPath, method string, handler gin.HandlerFunc, desc string, middlewares ...gin.HandlerFunc) {
	w.RegWithAPIType(rg, relPath, method, handler, desc, "", middlewares...)
}

func (w *wrapper) RegWithAPIType(rg *gin.RouterGroup, relPath, method string, handler gin.HandlerFunc, desc string, openApiType string, middlewares ...gin.HandlerFunc) {
	if w == nil {
		log.Panicf("wrapper nil")
	}
	absPath := path.Join(rg.BasePath(), relPath)
	// check w.tag
	if w.tag == "" {
		log.Panicf("wrapper tag empty cannot register [%v]%v", method, absPath)
	}
	// check method & path
	if w.paths.Get(absPath, method) != nil {
		log.Panicf("wrapper %v register [%v]%v already exist", w.tag, method, absPath)
	}
	// check route
	loaded, err := route.LoadOrStoreWithOpenAPIType(absPath, method, desc, w.permLevel, route.TagName{Tag: w.tag, Name: w.name}, openApiType, handler, middlewares...)
	if err != nil {
		log.Panicf("wrapper %v route err: %v", w.tag, err)
	}
	if loaded {
		return
	}
	// handler
	var handlers []gin.HandlerFunc
	handlers = append(handlers, w.middlewares...)
	switch w.permLevel {
	case route.PermNone:
	case route.PermNeedEnable:
	case route.PermNeedCheck:
	default:
		log.Panicf("wrapper %v permLevel %v unknown", w.tag, w.permLevel)
	}
	handlers = append(handlers, middlewares...)
	handlers = append(handlers, w.postMiddlewares...)
	handlers = append(handlers, handler)
	rg.Handle(method, relPath, handlers...)
}

func (w *wrapper) sub(tag string) *wrapper {
	var ret *wrapper
	for _, sub := range w.subs {
		if sub.tag == tag {
			return sub
		}
		if ret = sub.sub(tag); ret != nil {
			return ret
		}
	}
	return nil
}

func (w *wrapper) collectRoute() route.Route {
	ret := route.Route{
		Tag:  w.tag,
		Name: w.name,
	}
	for _, sub := range w.subs {
		if sub.permLevel >= route.PermNeedCheck {
			ret.Subs = append(ret.Subs, sub.collectRoute())
		}
	}
	return ret
}

func (w *wrapper) collectPerms(ret *[]route.TagName) {
	if w.permLevel >= route.PermNeedCheck {
		*ret = append(*ret, route.TagName{
			Tag:  w.tag,
			Name: w.name,
		})
	}
	for _, sub := range w.subs {
		sub.collectPerms(ret)
	}
}
