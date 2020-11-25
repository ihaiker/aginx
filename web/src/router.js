// import Vue from 'vue'
// import VueRouter from 'vue-router'

const AppContainer = () => import('@/containers/AppContainer');
const Nodes = () => import("@/views/Nodes")
const Servers = () => import('@/views/server/Server');
const ServerEdit = () => import('@/views/server/ServerEdit')
const Files = () => import('@/views/files/Files');
const FileEdit = () => import('@/views/files/FileEdit')
const Params = () => import("@/views/params/Param")
const BasicAuths = () => import("@/views/BasicAuths")
const Backup = () => import("@/views/Backup")
const Plugins = () => import("@/views/Plugins")
const Upstreams = () => import("@/views/upstream/Upstreams")
const UpstreamEdit = () => import("@/views/upstream/UpstreamEdit")
const Certs = () => import("@/views/Certs")
const Login = () => import("@/views/Login")

Vue.use(VueRouter);

export default new VueRouter({
    mode: 'hash',
    linkActiveClass: 'open active',
    scrollBehavior: () => ({y: 0}),
    routes: [
        {path: "/login", component: Login},
        {
            path: "/admin", component: AppContainer,
            children: [
                {path: "", redirect: "nodes"},
                {path: "nodes", component: Nodes},
                {path: "files", component: Files},
                {path: "file/edit", component: FileEdit},
                {path: "params", component: Params},
                {path: "servers", component: Servers},
                {path: "server/edit", component: ServerEdit},
                {path: "auths", component: BasicAuths},
                {path: "backup", component: Backup},
                {path: "plugins", component: Plugins},
                {path: "upstreams", component: Upstreams},
                {path: "upstream/edit", component: UpstreamEdit},
                {path: "certs", component: Certs},
                {path: '*', redirect: 'nodes'}
            ]
        },
        {path: '*', redirect: '/admin'}
    ]
})
