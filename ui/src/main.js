import Vue from 'vue'
import App from './App.vue'
import VueNativeSock from 'vue-native-websocket'

import store from './store'

Vue.use(VueNativeSock, 'ws://localhost:8089/ws', {store: store, format: 'json'});

// Components
import Navbar from './Nav.vue';
import LineChart from './LineChart.vue';
import HostList from './HostList.vue';
Vue.component('navbar', Navbar);
Vue.component('line-chart', LineChart);
Vue.component('host-list', HostList);

new Vue({
  el: '#app',
  store,
  render: h => h(App)
});


