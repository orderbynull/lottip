import Vue from 'vue'
import App from './App.vue'
import 'bulma/css/bulma.css'
import VueRouter from 'vue-router'
import MySQLTable from "./components/MySQLTable";
import RedisTable from "./components/RedisTable";

Vue.config.productionTip = false;
Vue.use(VueRouter);

const routes = [
    {path: '/mysql', component: MySQLTable},
    {path: '/redis', component: RedisTable},
];

const router = new VueRouter({
    routes
});

new Vue({
    router: router,
    render: h => h(App),
}).$mount('#app');
