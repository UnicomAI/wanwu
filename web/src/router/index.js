import Vue from 'vue'
import VueRouter from 'vue-router'
import { PERMS } from "./constants"
import { basePath } from "@/utils/config"

const routerPush = VueRouter.prototype.push;
VueRouter.prototype.push = function(location){
    return routerPush.call(this,location).catch(err=>{})
}

Vue.use(VueRouter)

let orgPermission = []

try{
    orgPermission = JSON.parse(localStorage.getItem('access_cert')).user.permission.orgPermission || []
}catch(e){
    // console.log(e)
    orgPermission = []
}

const constantRoutes = [
    {
        path: '/',
    },
    {
        path: '/doc',
        component:resolve =>require(['@/components/filePreview/DocPeview'],resolve),
    },
    {
        path: '/pdf',
        component:resolve =>require(['@/components/filePreview/PdfPreview'],resolve),
    },
    {
        path: '/pdfView',
        component:resolve =>require(['@/components/filePreview/pdfView'],resolve),
    },
    {
        path: '/txtView',
        component:resolve =>require(['@/components/filePreview/textView'],resolve),
    },
    {
        path: '/jsExcel',
        component:resolve =>require(['@/components/filePreview/JsPreviewExcel'],resolve),
    },
    {
        path: '/webChat/:id',
        component:resolve =>require(['@/views/agent'],resolve),
    },
    /* Temporarily remove Template Square public link */
    /*{
        path: '/public/templateSquare',
        component:resolve =>require(['@/views/templateSquare'],resolve),
    },
    {
        path: '/public/templateSquare/detail',
        component: resolve => require(['@/views/templateSquare/tempDetail.vue'],resolve),
    },*/
    {
        path: '/portal',
        name: 'portal',
        component:resolve =>require(['@/views/layout'],resolve),
        children:[
            {
                path: '/404',
                name: '404',
            },
            {
                path: '/userInfo',
                component:resolve =>require(['@/views/userCenter/components/info'],resolve),
            },
            {
                path: '/permission',
                component:resolve =>require(['@/views/permission'],resolve),
                meta:{perm: [PERMS.PERMISSION]},
            },
            {
                path: '/docCenter/pages/:id(.*)*',
                component:resolve => require(['@/views/docCenter'], resolve),
                // meta:{perm: [PERMS.DOC_CENTER]}
            },
            {
                path: '/modelAccess',
                component:resolve =>require(['@/views/modelAccess'],resolve),
                meta:{perm: [PERMS.MODEL]},
            },
            {
                path: '/tool',
                component:resolve => require(['@/views/tool'],resolve),
                meta:{perm: [PERMS.TOOL]},
            },
            {
                path: '/tool/detail/builtIn',
                component:resolve =>require(['@/views/tool/tool/builtIn/detail'],resolve),
                meta:{perm: [PERMS.TOOL]},
            },
            {
                path: '/tool/detail/custom',
                component:resolve =>require(['@/views/mcpManagementPublic/detail'],resolve),
                meta:{perm: [PERMS.TOOL]},
                props: { type: 'custom' }
            },
            {
                path: '/tool/detail/server',
                component:resolve =>require(['@/views/tool/mcp/server/detail'],resolve),
                meta:{perm: [PERMS.TOOL]},
            },
            {
                path: '/mcp',
                component:resolve => require(['@/views/mcpManagementPublic/square'],resolve),
                meta:{perm: [PERMS.MCP]},
            },
            {
                path: '/mcp/detail/square',
                component:resolve =>require(['@/views/mcpManagementPublic/detail'],resolve),
                meta:{perm: [PERMS.MCP]},
                props: { type: 'square' }
            },
            {
                path: '/explore',
                component:resolve =>require(['@/views/exploreSquare'],resolve),
                meta:{perm: [PERMS.EXPLORE]},
            },
            {
                path: '/explore/agent',
                component:resolve =>require(['@/views/agent'],resolve),
                meta:{perm: [PERMS.EXPLORE]},
            },
            {
                path: '/explore/workflow',
                component:resolve =>require(['@/views/workflowRun'],resolve),
                meta:{perm: [PERMS.EXPLORE]},
            },
            {
                path: '/explore/rag',
                component:resolve =>require(['@/views/rag'],resolve),
                meta:{perm: [PERMS.EXPLORE]},
            },
            {
                path: '/agent/test',
                component:resolve =>require(['@/views/agent/components/form'],resolve),
                meta:{perm: [PERMS.AGENT]},
            },
            {
                path: '/agent/templateDetail',
                name:'templateDetail',
                component:resolve =>require(['@/components/agentDetail'],resolve),
                meta:{perm: [PERMS.AGENT]},
            },
            {
                path: '/rag/test',
                component:resolve =>require(['@/views/rag/components/form'],resolve),
                meta:{perm: [PERMS.RAG]},
            },
            {
                path: '/workflow',
                component:resolve =>require(['@/views/workflowNew'],resolve),
                meta:{perm: [PERMS.WORKFLOW]},
            },
            {
                path: '/appSpace/:type',
                component:resolve =>require(['@/views/appSpace'],resolve),
                meta:{perm: [PERMS.RAG, PERMS.AGENT, PERMS.WORKFLOW]},
            },
            {
                path: '/knowledge',
                component:resolve =>require(['@/views/knowledge'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path: '/knowledge/doclist/:id',
                component:resolve =>require(['@/views/knowledge/component/doclist.vue'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
             {
                path: '/knowledge/fileUpload',
                component:resolve =>require(['@/views/knowledge/component/fileUpload.vue'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path: '/knowledge/section',
                component:resolve =>require(['@/views/knowledge/component/section.vue'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path: '/knowledge/keyword',
                component:resolve =>require(['@/views/knowledge/keyword'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path: '/knowledge/hitTest',
                component:resolve =>require(['@/views/knowledge/component/hitTest'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path: '/knowledge/communityReport',
                component:resolve =>require(['@/views/knowledge/component/communityReport'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path: '/knowledge/graphMap/:id',
                component:resolve =>require(['@/views/knowledge/component/graph'],resolve),
                meta:{perm: [PERMS.KNOWLEDGE]},
            },
            {
                path:'/safety',
                component:resolve =>require(['@/views/safety'],resolve),
                meta:{perm: [PERMS.SAFETY]},
            },
            {
                path:'/safety/wordList/:id',
                component:resolve =>require(['@/views/safety/component/wordList'],resolve),
                meta:{perm: [PERMS.SAFETY]},
            },
            {
                path:'/agent/publishSet',
                component:resolve =>require(['@/views/agent/web-URL'],resolve),
                meta:{perm: [PERMS.AGENT]},
            },
            {
                path:'/workflow/publishSet',
                component:resolve =>require(['@/views/agent/web-URL'],resolve),
                meta:{perm: [PERMS.WORKFLOW]},
            },
            {
                path:'/rag/publishSet',
                component:resolve =>require(['@/views/agent/web-URL'],resolve),
                meta:{perm: [PERMS.RAG]},
            },
            {
                path: '/templateSquare',
                component:resolve =>require(['@/views/templateSquare'],resolve),
            },
            {
                path: '/templateSquare/detail',
                component: resolve => require(['@/views/templateSquare/tempDetail.vue'],resolve),
            },
            {
                path: '/userCenter/*',
                name:'userCenter',
                component:resolve =>require(['@/views/userCenter'],resolve),
            },
        ]
    },
    {
        path: '/portal/*',
        name:'portalWithoutParams',
        component:resolve =>require(['@/views/layout'],resolve),
    },
    {
        path: '/portal/:path(.*)',
        name:'portalWithParams',
        component:resolve =>require(['@/views/layout'],resolve),
    },
    {
        path: '/login',
        component: () => import('@/views/auth/login'),
    },
    {
        path: '/register',
        component: () => import('@/views/auth/register'),
    },
    {
        path: '/reset',
        component: () => import('@/views/auth/reset'),
    },
    {
        path: '/oauth',
        component: () => import('@/views/auth/oauth'),
    },
    {
        path: '/:catchAll(.*)',
        redirect: "/"
    }
]


// CheckYesNoHasPermission
const hasPermission = (perm, route) => {
    if (!Array.isArray(perm)) return false
    if (route.meta && route.meta.perm) {
        return route.meta.perm.some(role => perm.includes(role))
    } else {
        return true
    }
}
// Combine HasPermission of Router restart
const filterAsyncRoutes = (routes, perm) => {
    const res = []

    routes.forEach(route => {
        const tmp = { ...route }
        if (hasPermission(perm, tmp)) {
            if (tmp.children) {
                tmp.children = filterAsyncRoutes(tmp.children, perm)
                if (tmp.children.length && !tmp.redirect) tmp.redirect = tmp.children[0].path
            }
            res.push(tmp)
        }
    })
    return res
}

const baseConfig = {
    mode: 'history',
    base: basePath + '/aibase',//process.env.BASE_URL,
    scrollBehavior (to, from, savedPosition) {
        return { x: 0, y: 0 }
    },
}

let router = new VueRouter({
    ...baseConfig,
    routes: filterAsyncRoutes(constantRoutes, orgPermission)
})

export const replaceRouter = (permission) => {
    // Create new Router instance
    const newRouter = new VueRouter({
        ...baseConfig,
        routes: filterAsyncRoutes(constantRoutes, permission), // Use new RouteConfiguration
    })

    // Replace existing Router
    router.matcher = newRouter.matcher
}

export default router
