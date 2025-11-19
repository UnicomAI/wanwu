import { PERMS } from "@/router/permission"

export const menuList = [
    {
        name: 'menu.modelAccess',
        key: 'modelAccess',
        img: require('@/assets/imgs/model.svg'),
        imgActive: require('@/assets/imgs/model_active.svg'),
        path: '/modelAccess',
        perm: PERMS.MODEL,
    },
    {
        name: 'menu.knowledge',
        key: 'knowledge',
        img: require('@/assets/imgs/knowledge.svg'),
        imgActive: require('@/assets/imgs/knowledge_active.svg'),
        path: '/knowledge',
        perm: PERMS.KNOWLEDGE,
    },
    {
        name: 'menu.tool',
        key: 'tool',
        img: require('@/assets/imgs/tool.svg'),
        imgActive: require('@/assets/imgs/tool_active.svg'),
        path: '/tool',
        perm: PERMS.TOOL,
    },
    {
        name: 'menu.safetyGuard',
        key: 'safetyGuard',
        img: require('@/assets/imgs/safety.svg'),
        imgActive: require('@/assets/imgs/safety_active.svg'),
        path: '/safety',
        perm: PERMS.SAFETY,
    },
    {
        key: 'line',
        perm: [PERMS.MODEL, PERMS.KNOWLEDGE, PERMS.TOOL]
    },
    {
        name: 'menu.app.rag',
        key: 'rag',
        img: require('@/assets/imgs/rag.svg'),
        imgActive: require('@/assets/imgs/rag_active.svg'),
        path: '/appSpace/rag',
        perm: PERMS.RAG
    },
    {
        name: 'menu.app.workflow',
        key: 'workflow',
        img: require('@/assets/imgs/workflow_icon.svg'),
        imgActive: require('@/assets/imgs/workflow_icon_active.svg'),
        path: '/appSpace/workflow',
        perm: PERMS.WORKFLOW
    },
    {
        name: 'menu.app.agent',
        key: 'agent',
        img: require('@/assets/imgs/agent.svg'),
        imgActive: require('@/assets/imgs/agent_active.svg'),
        path: '/appSpace/agent',
        perm: PERMS.AGENT
    },
    {
        key: 'line',
        perm: [PERMS.RAG, PERMS.WORKFLOW, PERMS.AGENT]
    },
    {
        name: 'menu.mcp',
        key: 'mcpManage',
        img: require('@/assets/imgs/mcp_menu.svg'),
        imgActive: require('@/assets/imgs/mcp_menu_active.svg'),
        path: '/mcp',
        perm: PERMS.MCP,
    },
    {
        name: 'menu.explore',
        key: 'explore',
        img: require('@/assets/imgs/explore.svg'),
        imgActive: require('@/assets/imgs/explore_active.svg'),
        path: '/explore',
        perm: PERMS.EXPLORE
    },
    {
        name: 'menu.templateSquare',
        key: 'templateSquare',
        img: require('@/assets/imgs/template_square.svg'),
        imgActive: require('@/assets/imgs/template_square_active.svg'),
        path: '/templateSquare',
    },
]