import Vue from 'vue'
import App from './App.vue'
import Buefy from 'buefy'
import 'buefy/lib/buefy.css'
import VueNativeSock from 'vue-native-websocket'

Vue.use(Buefy);
import store from './store'

Vue.use(VueNativeSock, 'ws://localhost:8089/ws', {store: store, format: 'json'});

// Components
import HostList from './HostList.vue';
import OverallStats from './OverallStats.vue';
Vue.component('host-list', HostList);
Vue.component('overall-stats', OverallStats);

new Vue({
  el: '#app',
  store,
  render: h => h(App)
});


