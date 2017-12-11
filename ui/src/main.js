import Vue from 'vue'
import App from './App.vue'
import VueNativeSock from 'vue-native-websocket'
import VueResource from 'vue-resource';

import store from './store'

Vue.use(VueNativeSock, 'ws://localhost:8089/ws', {store: store, format: 'json'});
Vue.use(VueResource);

// Components
import Navbar from './shared/Nav.vue';
import LineChart from './shared/LineChart.vue';
import BarChart from './shared/BarChart.vue';
import Dashboard from './shared/Dashboard.vue';
Vue.component('navbar', Navbar);
Vue.component('dashboard', Dashboard);
Vue.component('line-chart', LineChart);
Vue.component('bar-chart', BarChart);

new Vue({
  el: '#app',
  store,
  render: h => h(App)
});


