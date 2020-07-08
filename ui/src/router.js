// import Vue from 'vue'
// import VueRouter from 'vue-router'

const AppContainer = () => import('@/containers/AppContainer');
const Login = () => import('@/views/Login');

const Servers= () => import('@/views/config/Server');
const Files = () => import('@/views/files/Files');

Vue.use(VueRouter);

export default new VueRouter({
    mode: 'hash',
    linkActiveClass: 'open active',
    scrollBehavior: () => ({y: 0}),
    routes: [
        {path: "/signin", name: 'signin', component: Login},
        {
            path: "/admin", component: AppContainer,
            children: [
                {path: "", redirect: "files"},
                {path: "files", component: Files},
                {path: "server", component: Servers},
                {path: '*', redirect: 'files'}
            ]
        },
        {path: '*', redirect: '/admin'}
    ]
})
