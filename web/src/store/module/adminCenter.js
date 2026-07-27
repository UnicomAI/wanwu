const clone = value => JSON.parse(JSON.stringify(value));

export const adminCenter = {
  namespaced: true,
  state: {
    activeList: null,
  },
  mutations: {
    SET_ACTIVE_LIST(state, payload) {
      state.activeList = clone(payload);
    },
    CLEAR_ACTIVE_LIST(state) {
      state.activeList = null;
    },
  },
  getters: {
    activeList: state => state.activeList,
  },
};
