import Vue from 'vue';
import VueRouter from 'vue-router';
import { PERMS } from './constants';
import { basePath } from '@/utils/config';
import { PROMPT, MCP, TOOL } from '@/views/tool/constants';

const routerPush = VueRouter.prototype.push;
VueRouter.prototype.push = function (location) {
  return routerPush.call(this, location).catch(err => {});
};

Vue.use(VueRouter);

let orgPermission = [];

try {
  orgPermission =
    JSON.parse(localStorage.getItem('access_cert')).user.permission
      .orgPermission || [];
} catch (e) {
  // console.log(e)
  orgPermission = [];
}

const constantRoutes = [
  {
    path: '/',
  },
  {
    path: '/doc',
    component: resolve =>
      require(['@/components/filePreview/DocPeview'], resolve),
  },
  {
    path: '/txtView',
    component: resolve =>
      require(['@/components/filePreview/textView'], resolve),
  },
  {
    path: '/jsExcel',
    component: resolve =>
      require(['@/components/filePreview/JsPreviewExcel'], resolve),
  },
  {
    path: '/webChat/:id',
    component: resolve => require(['@/views/agent'], resolve),
  },
  /* 暂时去掉模板广场公网的链接 */
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
    component: resolve => require(['@/views/layout'], resolve),
    children: [
      {
        path: '/404',
        name: '404',
      },
      {
        path: '/userInfo',
        component: resolve =>
          require(['@/views/userCenter/components/info'], resolve),
      },
      {
        path: '/permission',
        component: resolve => require(['@/views/permission'], resolve),
        meta: { perm: [PERMS.ADMIN_CENTER] },
      },
      {
        path: '/operation',
        component: resolve => require(['@/views/operation'], resolve),
        meta: { perm: [PERMS.OPERATION] },
      },
      {
        path: '/docCenter/pages/:id(.*)*',
        component: resolve => require(['@/views/docCenter'], resolve),
      },
      {
        path: '/aiAssistant',
        component: resolve => require(['@/views/aiAssistant/index'], resolve),
        meta: { perm: [PERMS.WGA_OPENCLAW] },
      },
      {
        path: '/modelAccess',
        component: resolve => require(['@/views/modelAccess'], resolve),
        meta: { perm: [PERMS.MODEL_MANAGE] },
      },
      {
        path: '/modelAccess/modelExprience',
        component: resolve =>
          require(['@/views/modelExprience/index'], resolve),
        meta: { perm: [PERMS.MODEL_MANAGE] },
      },
      {
        path: '/skill',
        component: resolve =>
          require(['@/views/templateSquare/skills/index.vue'], resolve),
        meta: { perm: [PERMS.SKILL] },
      },
      {
        path: '/skill/detail',
        component: resolve =>
          require(['@/views/templateSquare/skills/detail.vue'], resolve),
        meta: { perm: [PERMS.SKILL] },
      },
      {
        path: '/skill/workshop',
        component: resolve =>
          require(['@/views/templateSquare/skills/skillWorkshop.vue'], resolve),
        meta: { perm: [PERMS.SKILL] },
      },
      {
        path: '/skill/publishConfig',
        component: () =>
          import('@/views/templateSquare/skills/custom/publishConfig/index.vue'),
        meta: { perm: [PERMS.SKILL] },
      },
      {
        path: '/tool',
        component: resolve => require(['@/views/tool'], resolve),
        meta: { perm: [PERMS.TOOL], routeType: TOOL, cacheName: 'Tool' },
      },
      {
        path: '/tool/detail/builtIn',
        component: resolve =>
          require(['@/views/tool/tool/builtIn/detail'], resolve),
        meta: { perm: [PERMS.TOOL] },
      },
      {
        path: '/prompt',
        component: resolve => require(['@/views/tool'], resolve),
        meta: { perm: [PERMS.PROMPT], routeType: PROMPT, cacheName: 'Tool' },
      },
      {
        path: '/promptEvaluate',
        name: 'promptEvaluate',
        meta: { perm: [PERMS.PROMPT] },
        component: resolve => require(['@/views/promptEvaluate'], resolve),
      },
      {
        path: '/mcpService',
        component: resolve => require(['@/views/tool'], resolve),
        meta: { perm: [PERMS.MCP_SERVICE], routeType: MCP, cacheName: 'Tool' },
      },
      {
        path: '/mcpService/detail/custom',
        component: resolve =>
          require(['@/views/mcpManagementPublic/detail'], resolve),
        meta: { perm: [PERMS.MCP_SERVICE] },
        props: { type: 'custom' },
      },
      {
        path: '/mcpService/detail/server',
        component: resolve =>
          require(['@/views/tool/mcp/server/detail'], resolve),
        meta: { perm: [PERMS.MCP_SERVICE] },
      },
      {
        path: '/mcp',
        component: resolve =>
          require(['@/views/mcpManagementPublic/square'], resolve),
        meta: { perm: [PERMS.MCP], cacheName: 'McpSquare' },
      },
      {
        path: '/mcp/detail/square',
        component: resolve =>
          require(['@/views/mcpManagementPublic/detail'], resolve),
        meta: { perm: [PERMS.MCP] },
        props: { type: 'square' },
      },
      {
        path: '/explore',
        component: resolve => require(['@/views/exploreSquare'], resolve),
        meta: { perm: [PERMS.EXPLORE] },
      },
      {
        path: '/explore/agent',
        component: resolve => require(['@/views/agent'], resolve),
        meta: { perm: [PERMS.EXPLORE] },
      },
      {
        path: '/explore/workflow',
        component: resolve => require(['@/views/workflowRunNew'], resolve),
        meta: { perm: [PERMS.EXPLORE] },
      },
      {
        path: '/explore/rag',
        component: resolve => require(['@/views/rag'], resolve),
        meta: { perm: [PERMS.EXPLORE] },
      },
      {
        path: '/agent/test',
        component: resolve =>
          require(['@/views/agent/components/form'], resolve),
        meta: { perm: [PERMS.AGENT] },
      },
      {
        path: '/agent/promptCompare/:id',
        component: resolve =>
          require(['@/views/agent/components/prompt/compare'], resolve),
        meta: { perm: [PERMS.AGENT] },
      },
      {
        path: '/rag/test',
        component: resolve => require(['@/views/rag/components/form'], resolve),
        meta: { perm: [PERMS.RAG] },
      },
      {
        path: '/workflow',
        component: resolve => require(['@/views/workflowNew'], resolve),
        meta: { perm: [PERMS.WORKFLOW] },
      },
      {
        path: '/appSpace/:type',
        component: resolve => require(['@/views/appSpace'], resolve),
        meta: { perm: [PERMS.RAG, PERMS.AGENT, PERMS.WORKFLOW] },
      },
      {
        path: '/knowledge',
        component: resolve => require(['@/views/knowledge'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE], cacheName: 'Knowledge' },
      },
      {
        path: '/knowledge/doclist/:id',
        component: resolve =>
          require(['@/views/knowledge/knowledgeDatabase/doclist.vue'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE], cacheName: 'KnowledgeDoclist' },
      },
      {
        path: '/knowledge/qa/docList/:id',
        component: resolve =>
          require(['@/views/knowledge/qaDatabase/docList.vue'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE], cacheName: 'KnowledgeQaDoclist' },
      },
      // {
      //   path: '/knowledge/db/docList/:id',
      //   component: resolve =>
      //     require(['@/views/knowledge/dbDatabase/docList.vue'], resolve),
      //   meta: { perm: [PERMS.KNOWLEDGE] },
      // },
      {
        path: '/knowledge/fileUpload',
        component: resolve =>
          require([
            '@/views/knowledge/knowledgeDatabase/fileUpload.vue',
          ], resolve),
        meta: { perm: [PERMS.KNOWLEDGE] },
      },
      {
        path: '/knowledge/section',
        component: resolve =>
          require(['@/views/knowledge/component/section.vue'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE] },
      },
      {
        path: '/knowledge/keyword',
        component: resolve => require(['@/views/knowledge/keyword'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE] },
      },
      {
        path: '/knowledge/hitTest',
        component: resolve =>
          require(['@/views/knowledge/component/hitTest'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE] },
      },
      {
        path: '/knowledge/communityReport',
        component: resolve =>
          require(['@/views/knowledge/component/communityReport'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE] },
      },
      {
        path: '/knowledge/graphMap/:id',
        component: resolve =>
          require(['@/views/knowledge/component/graph'], resolve),
        meta: { perm: [PERMS.KNOWLEDGE] },
      },
      {
        path: '/safety',
        component: resolve => require(['@/views/safety'], resolve),
        meta: { perm: [PERMS.SAFETY] },
      },
      {
        path: '/safety/wordList/:id',
        component: resolve =>
          require(['@/views/safety/component/wordList'], resolve),
        meta: { perm: [PERMS.SAFETY] },
      },
      {
        path: '/agent/publishSet',
        component: resolve => require(['@/components/publishConfig'], resolve),
        meta: { perm: [PERMS.AGENT] },
      },
      {
        path: '/generalAgent',
        component: resolve =>
          require(['@/views/generalAgent/index.vue'], resolve),
        meta: { perm: [PERMS.WGA_WANWU_BOT] },
      },
      {
        path: '/generalAgent/skills/publishConfig',
        component: () =>
          import('@/views/templateSquare/skills/custom/publishConfig/index.vue'),
        meta: { perm: [PERMS.WGA_WANWU_BOT] },
      },
      {
        path: '/workflow/publishSet',
        component: resolve => require(['@/components/publishConfig'], resolve),
        meta: { perm: [PERMS.WORKFLOW] },
      },
      {
        path: '/rag/publishSet',
        component: resolve => require(['@/components/publishConfig'], resolve),
        meta: { perm: [PERMS.RAG] },
      },
      {
        path: '/statisticsDashboard',
        component: resolve => require(['@/views/statisticsDashboard'], resolve),
        meta: { perm: [PERMS.OBSERVATION_STATISTIC] },
      },
      {
        path: '/openApiKey',
        component: resolve => require(['@/views/apiKeyManagement'], resolve),
        meta: { perm: [PERMS.API_KEY_MANAGE] },
      },
      {
        path: '/templateSquare',
        component: resolve => require(['@/views/templateSquare'], resolve),
      },
      {
        path: '/templateSquare/detail',
        component: resolve =>
          require(['@/views/templateSquare/tempDetail.vue'], resolve),
      },
      {
        path: '/skillSquare',
        component: resolve => require(['@/views/skillSquare'], resolve),
        meta: { perm: [PERMS.SKILL_SQUARE] },
      },
      {
        path: '/skillSquare/detail',
        component: resolve =>
          require(['@/views/skillSquare/detail.vue'], resolve),
        meta: { perm: [PERMS.SKILL_SQUARE] },
      },
      {
        path: '/userCenter/*',
        name: 'userCenter',
        component: resolve => require(['@/views/userCenter'], resolve),
      },
    ],
  },
  {
    path: '/portal/*',
    name: 'portalWithoutParams',
    component: resolve => require(['@/views/layout'], resolve),
  },
  {
    path: '/portal/:path(.*)',
    name: 'portalWithParams',
    component: resolve => require(['@/views/layout'], resolve),
  },
  {
    path: '/auth',
    component: () => import('@/views/auth/layout'),
    children: [
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
    ],
  },
  {
    path: '/oauth',
    component: () => import('@/views/auth/oauth'),
  },
  {
    path: '/:catchAll(.*)',
    redirect: '/',
  },
];

// 判断是否有权限
const hasPermission = (perm, route) => {
  if (!Array.isArray(perm)) return false;
  if (route.meta?.perm) {
    return route.meta.perm.some(role => perm.includes(role));
  } else {
    return true;
  }
};
// 把有权限的路由重新组合
const filterAsyncRoutes = (routes, perm) => {
  const res = [];

  routes.forEach(route => {
    const tmp = { ...route };
    if (hasPermission(perm, tmp)) {
      if (tmp.children) {
        tmp.children = filterAsyncRoutes(tmp.children, perm);
        if (tmp.children.length && !tmp.redirect)
          tmp.redirect = tmp.children[0].path;
      }
      res.push(tmp);
    }
  });
  return res;
};

const baseConfig = {
  mode: 'history',
  base: basePath + '/aibase', //process.env.BASE_URL,
  scrollBehavior(to, from, savedPosition) {
    return { x: 0, y: 0 };
  },
};

const router = new VueRouter({
  ...baseConfig,
  routes: filterAsyncRoutes(constantRoutes, orgPermission),
});

export const replaceRouter = permission => {
  // 创建新的 Router 实例
  const newRouter = new VueRouter({
    ...baseConfig,
    routes: filterAsyncRoutes(constantRoutes, permission), // 使用新的路由配置
  });

  // 替换现有的路由器
  router.matcher = newRouter.matcher;
};

export default router;
