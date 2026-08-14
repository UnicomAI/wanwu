export const layout = {
  namespaced: true,
  state: {
    aboutDialogVisible: false,
  },
  mutations: {
    OPEN_ABOUT_DIALOG(state) {
      state.aboutDialogVisible = true;
    },
    CLOSE_ABOUT_DIALOG(state) {
      state.aboutDialogVisible = false;
    },
  },
  getters: {
    aboutDialogVisible: state => state.aboutDialogVisible,
  },
};
