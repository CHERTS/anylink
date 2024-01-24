import Vue from 'vue'
import App from './App.vue'
import './plugins/element'
import "./plugins/mixin";
import request from './plugins/request'
import router from "./plugins/router";

//TODO
Vue.config.productionTip = false

const vm = new Vue({
    data: {
        // Determine whether to log in
        isLogin: false,
    },
    router,
    render: h => h(App),
}).$mount('#app')

request(vm)