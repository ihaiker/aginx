<template>
    <div class="app">
        <Loadings/>
        <AppHeader fixed>
            <SidebarToggler class="d-lg-none" display="md" mobile/>
            <b-link class="navbar-brand" to="/">
                <img class="navbar-brand-full" src="@/assets/images/logo.png" width="89" height="25" alt="Sudis Logo">
                <img class="navbar-brand-minimized" src="/favicon.ico" width="30" height="30" alt="Sudis Logo">
            </b-link>
            <SidebarToggler class="d-md-down-none" display="lg"/>
            <b-navbar-nav class="d-md-down-none">
                <b-nav-item class="px-3" to="/">Nginx配置管理程序
                    <span class="text-danger font-weight-bold">【当前节点：{{ $store.getters.node.name }}】</span>
                </b-nav-item>
            </b-navbar-nav>
            <b-navbar-nav class="ml-auto d-md-down-none">
                <b-nav-item class="px-3" href="https://github.com/ihaiker/aginx/issues/new" target="_blank">/提交BUG/
                </b-nav-item>
                <b-nav-item class="px-3" href="https://github.com/ihaiker/aginx" target="_blank">/源码/</b-nav-item>
                <!--
                <b-nav-item>
                    <i class="icon-bell"/>
                    <b-badge pill variant="danger">5</b-badge>
                </b-nav-item>
                -->
            </b-navbar-nav>
            <!--
                <AsideToggler class="d-none d-lg-block"/>
                <AsideToggler class="d-lg-none" mobile/>
            -->
        </AppHeader>
        <div class="app-body">
            <Sidebar fixed>
                <SidebarHeader/>
                <SidebarForm/>
                <SidebarNav :navItems="nav"/>
                <SidebarFooter/>
                <SidebarMinimizer/>
            </Sidebar>
            <main class="main">
                <router-view/>
            </main>
            <Aside fixed>
                <AppAside/>
            </Aside>
        </div>
        <AppFooter>
            <div></div>
            <div class="ml-auto">
                Aginx
                <span class="ml-1">&copy; 2019.</span>
            </div>
            <div class="ml-auto">
                <span class="mr-1">Powered by</span>
                <a href="http://shui.renzhen.la" target="_blank">Haiker</a>
            </div>
        </AppFooter>
    </div>
</template>
<style>
.app-footer, .sidebar .sidebar-minimizer {
    flex: 0 0 30px;
}

.sidebar .sidebar-minimizer::before {
    width: 30px;
    height: 30px;
}

.sidebar-minimized .sidebar .sidebar-minimizer {
    width: 50px;
    height: 30px;
}
</style>
<script>
import {
    Aside,
    AsideToggler,
    Footer as AppFooter,
    Header as AppHeader,
    HeaderDropdown as AppHeaderDropdown,
    Sidebar,
    SidebarFooter,
    SidebarForm,
    SidebarHeader,
    SidebarMinimizer,
    SidebarNav,
    SidebarToggler
} from '@coreui/vue'

import AppAside from './AppAside'
import Loadings from "../plugins/loadings";

export default {
    name: 'AppContainer',
    components: {
        Loadings, AppHeader, AppFooter,
        Aside, AsideToggler, AppAside, Sidebar, SidebarForm, SidebarFooter, SidebarToggler,
        SidebarHeader, SidebarNav, SidebarMinimizer, AppHeaderDropdown
    },
    methods: {
    },
    computed: {
        node() {
            let nodeName = localStorage.getItem('aginx.node.name');
            if (nodeName) {
                return nodeName
            }
            return ""
        },
    },
    data: () => ({
        nav: [
            {
                name: '节点选择',
                url: '/admin/nodes',
                icon: 'icons cui-puzzle',
                badge: {
                    variant: 'primary',
                    text: '新'
                }
            },
            {
                name: '文件管理',
                url: '/admin/files',
                icon: 'fa fa-file-o',
            },
            {
                name: '参数列表',
                url: '/admin/params',
                icon: 'icons cui-list',
            },
            {
                name: '反向代理(server)',
                url: '/admin/servers',
                icon: "fa fa-globe",
            },
            {
                name: '负载均衡(upstream)',
                url: '/admin/upstreams',
                icon: "icons icon-organization",
            },
            {
                name: '证书管理',
                url: '/admin/certs',
                icon: "icons icon-key",
            },
            {
                name: 'Basic认证管理',
                url: '/admin/auths',
                icon: "icons icon-lock",
            },
            {
                name: '备份管理',
                url: '/admin/backup',
                icon: "icons cui-cloud-download",
            },
            {
                name: '插件仓库',
                url: '/admin/plugins',
                icon: "icons cui-puzzle",
            }
        ]
    }),
}
</script>
