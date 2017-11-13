import Vuex from 'vuex'
import Vue from 'vue';

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    hostList: {name: 'la'},
    socket: {
      message: '',
      isConnected: false,
    }
  },
  mutations: {
    SOCKET_ONOPEN(state, event) {
      state.socket.isConnected = true
    },
    SOCKET_ONCLOSE(state, event) {
      state.socket.isConnected = false
    },
    SOCKET_ONERROR(state, event) {
      console.error(state, event)
    },
    // default handler called for all methods
    SOCKET_ONMESSAGE(state, message) {
      state.socket.message = message.data
    },
    hostList(state, host) {
      state.hostList = host;
    }
  }
});

