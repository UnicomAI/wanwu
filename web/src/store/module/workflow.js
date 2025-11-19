export const workflow = {
    namespaced: true,
    state: {
        nodeIdMap:{}, //用来在选择ReferenceTypeData when ，根据节点id回显节点Name
        activeName: ['1', '2'],
        lastDebugResult:{},
        knowledgeData:[]
    },

    mutations: {
        SET_NODEID_MAP(state, data) {
            state.nodeIdMap = { ...state.nodeIdMap, ...data }
            window.localStorage.setItem('nodeIdMap',JSON.stringify(state.nodeIdMap))
        },
        SET_ACTIVENAME(state,data){
            state.activeName = data
        },
        SET_DEBUG_STATE(state,data){
            state.lastDebugResult = data
        },
        SET_KNOWLEDGE_LIST(state,data){
            state.knowledgeData = data
        },
    },
    actions: {
        setNodeIdMap({ commit },data) {
            commit('SET_NODEID_MAP',data)
        },
        setactiveName({ commit },data){
            commit('SET_ACTIVENAME',data)
        },
        setLastDebugResult({ commit },data){
            console.log('setLastDebugResult:',data)
            commit('SET_DEBUG_STATE',data)
        },
        setKnowledgeList({ commit },data){
            commit('SET_KNOWLEDGE_LIST',data)
        },
    },
    getters: {
        nodeIdMap(state) {
            return state.nodeIdMap
        },
        activeName(state) {
            return state.activeName
        },
        lastDebugResult(state){
            return state.lastDebugResult
        },
        knowledgeData(state){
            return state.knowledgeData
        },
    }
}
