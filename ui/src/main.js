import Vue from 'vue'
import App from './App.vue'
import Buefy from 'buefy'
import 'buefy/lib/buefy.css'
import VueNativeSock from 'vue-native-websocket'

import store from './store'

Vue.use(VueNativeSock, 'ws://localhost:8089/ws', {store: store, format: 'json'});
Vue.use(Buefy);

// Components
import HostList from './HostList.vue';
Vue.component('host-list', HostList);

new Vue({
  el: '#app',
  store,
  render: h => h(App)
});


