import Vue from 'vue';
import Vuex from 'vuex';
import VuexPersistence from 'vuex-persist';
import { user } from './module/user';
import { app } from './module/app';
import { workflow } from './module/workflow';
import { adminCenter } from './module/adminCenter';
import { layout } from './module/layout';

Vue.use(Vuex);
// 用户信息持久化
const vuexLocal = new VuexPersistence({
  key: 'access_cert',
  storage: localStorage,
  modules: ['user'],
});

export const store = new Vuex.Store({
  modules: {
    user,
    app,
    workflow,
    adminCenter,
    layout,
  },
  plugins: [vuexLocal.plugin],
});
