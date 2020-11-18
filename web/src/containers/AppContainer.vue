<template>
    <div class="app">
        <Loadings/>
        <AppHeader fixed>
            <SidebarToggler class="d-lg-none" display="md" mobile/>
            <b-link class="navbar-brand" to="/admin">
                <img class="navbar-brand-full" src="@/assets/images/logo.png" width="89" height="25" alt="Sudis Logo">
                <img class="navbar-brand-minimized" src="/favicon.ico" width="30" height="30" alt="Sudis Logo">
            </b-link>

            <SidebarToggler class="d-md-down-none" display="lg"/>

            <b-navbar-nav class="d-md-down-none">
                <b-nav-item class="px-3" to="/admin">Nginx配置管理程序</b-nav-item>
            </b-navbar-nav>

            <b-navbar-nav class="ml-auto d-md-down-none">
                <b-nav-item class="px-3" href="https://github.com/ihaiker/aginx/issues/new" target="_blank">/提交BUG/
                </b-nav-item>
                <b-nav-item class="px-3" href="https://github.com/ihaiker/aginx" target="_blank">/源码下载/</b-nav-item>

                <AppHeaderDropdown right no-caret class="mr-3">
                    <template slot="header">
                        <img src="@/assets/images/logo2.png" class="img-avatar"/>
                    </template>
                    <template slot="dropdown">
                        <b-dropdown-item><i class="fa fa-user-circle"/> <strong>{{userName}}</strong></b-dropdown-item>
                        <!--<b-dropdown-item><i class="fa fa-shield"/> 修改密码</b-dropdown-item> -->
                        <b-dropdown-item @click="logout"><i class="fa fa-lock"/> 退出登录</b-dropdown-item>
                    </template>
                </AppHeaderDropdown>

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
    .sidebar .sidebar-minimizer::before{
        width: 30px;
        height: 30px;
    }
    .sidebar-minimized .sidebar .sidebar-minimizer {
        width: 50px;
        height: 30px;
    }
</style>
<script>
    import {Header as AppHeader, Footer as AppFooter} from '@coreui/vue'
    import {
        Sidebar, SidebarFooter, SidebarForm,
        SidebarHeader, SidebarMinimizer, SidebarNav, SidebarToggler
    } from '@coreui/vue'
    import {Aside, AsideToggler} from '@coreui/vue'
    import {HeaderDropdown as AppHeaderDropdown} from '@coreui/vue'

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
            logout() {
                localStorage.removeItem("token");
                this.$router.push({path: "/signin"});
            }
        },
        mounted() {
            let token = localStorage.getItem("token")
            if (token === null || token === undefined || token === "") {
                this.logout()
            }
        },
        computed: {
            userName() {
                return localStorage.getItem("x-user");
            }
        },
        data: () => ({
            nav: [
                {
                    name: '文件方式管理',
                    url: '/admin/files',
                    icon: 'icons cui-file',
                    badge: {
                        variant: 'primary',
                        text: '新'
                    }
                },
                {
                    name: '服务管理',
                    url: '/admin/server',
                    icon: "icons cui-puzzle",
                }
            ]
        })
    }
</script>
