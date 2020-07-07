// import Vue from 'vue'
// import VueRouter from 'vue-router'

const AppContainer = () => import('@/containers/AppContainer');
const Login = () => import('@/views/Login');

const Config = () => import('@/views/config/Config');
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
                {path: "", redirect: "config"},
                {path: "files", component: Files},
                {path: "config", component: Config},
                {path: '*', redirect: 'config'}
            ]
        },
        {path: '*', redirect: '/admin'}
    ]
})
