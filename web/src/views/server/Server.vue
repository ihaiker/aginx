<template>
    <div>
        <v-title title="代理列表" title-class="icons cui-puzzle">
            <router-link to="/admin/server/edit">
                <i class="fa fa-plus-circle"></i> 添加服务
            </router-link>
        </v-title>
        <div class="p-3">
            <div class="form-group form-inline">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text">域名</span>
                    </div>
                    <input class="form-control" v-model="searchName" type="text" placeholder="域名"
                           @keyup.enter="queryServices">
                    <div class="input-group-append">
                        <button class="btn btn-primary" @click="queryServices">
                            <i class="fa fa-search-plus"></i> 搜索
                        </button>
                    </div>
                </div>
            </div>

            <table class="table table-bordered table-hover">
                <thead>
                <tr>
                    <th scope="col" style="width: 240px;" class="text-wrap">协议/域名/描述</th>
                    <th scope="col" style="width: 100px;">监听</th>
                    <th scope="col">代理地址</th>
                    <th scope="col" style="width: 160px;">操作</th>
                </tr>
                </thead>
                <tbody>
                <template v-for="(server,idx) in services">
                    <tr v-if="showPage(idx)">
                        <td>
                            <span class="badge badge-dark">{{ server.protocol }}</span>
                            <span v-for="d in server.domains" class="text-success font-weight-bold ml-2">
                                {{ d }}
                            </span>
                            <div class="text-black-50" v-if="server.commit">
                                {{ server.commit }}
                            </div>
                        </td>
                        <td>
                            <div v-for="l in server.listens">
                                <span>{{ l.host }}:{{ l.port }}</span>
                                <span v-if="l.default" class="badge badge-success ml-2">默认</span>
                                <span v-if="l.http2" class="badge badge-info ml-2">http2</span>
                                <span v-if="l.ssl" class="badge badge-danger ml-2">ssl</span>
                            </div>
                        </td>
                        <td>
                            <div v-if="server.protocol !== 'http'">
                                转向负载：
                                <router-link :to="{path:'/admin/upstream/edit',query:{name:server.proxyPass}}"
                                             class="text-primary font-weight-bold">
                                    {{ server.proxyPass }}
                                </router-link>
                            </div>
                            <div v-if="server.protocol === 'http'" v-for="(loc) in server.locations">
                                <div class="badge badge-light">
                                    {{ loc.path }}
                                </div>
                                <template v-if="loc.type === 'html'">
                                    静态文件
                                    <span class="text-success font-weight-bold">
                                        {{ loc.html.model }}: {{ loc.html.path }}
                                    </span>
                                </template>
                                <template v-else-if="loc.type === 'upstream'">
                                    负载均衡
                                    <router-link :to="{path:'/admin/upstream/edit',query:{name:loc.upstream.name}}"
                                                 class="text-primary font-weight-bold">
                                        {{ loc.upstream.name }}
                                    </router-link>
                                    {{ loc.upstream.path }}
                                </template>
                                <template v-else-if="loc.type === 'http'">
                                    动态代理
                                    <span class="text-primary font-weight-bold">
                                        {{ loc.http.to }}
                                    </span>
                                </template>
                                <template v-else>
                                    用户定义
                                </template>
                                <span v-if="loc.commit" class="text-secondary">{{ loc.commit }}</span>
                            </div>
                        </td>
                        <td>
                            <div class="d-flex justify-content-around">
                                <button @click="editServer(server)" class="btn btn-sm btn-outline-primary">
                                    <i class="fa fa-edit"></i> 编辑
                                </button>
                                <button @click="deleteServer(server.queries)" class="btn btn-sm btn-outline-danger">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
                        </td>
                    </tr>
                </template>
                </tbody>
                <tfoot>
                <tr>
                    <td colspan="4">
                        <x-page :items="page" @change="page.page = $event"/>
                    </td>
                </tr>
                </tfoot>
            </table>
        </div>
    </div>
</template>
<style>
.badge {
    font-size: 14px;
}
</style>
<script>
import VTitle from "../../plugins/vTitle";
import VueInputAutowidth from 'vue-input-autowidth'
import Delete from "../../plugins/delete";
import XPage from "@/plugins/XPage";

export default {
    name: "Files",
    components: {XPage, Delete, VTitle, VueInputAutowidth},
    data: () => ({
        services: [],
        searchName: "",
        page: {
            page: 1, total: 0, limit: 12,
        }
    }),
    mounted() {
        this.queryServices();
    },
    methods: {
        showPage(idx) {
            return idx >= (this.page.page - 1) * this.page.limit
                && idx < (this.page.page * this.page.limit)
        },
        refresh() {
            this.queryServices();
        },
        queryServices() {
            this.startLoading();
            this.page.page = 1;
            let self = this;
            let url = "/admin/api/server";
            if (this.searchName !== "") {
                url += "?name=" + encodeURI(this.searchName);
            }
            this.$axios.get(url).then(res => {
                self.services = res;
                self.page.total = self.services.length;
            }).catch(e => {
                self.$toast.error(e.message);
            }).finally(() => {
                self.finishLoading()
            })
        },
        editServer(server) {
            if (server.protocol !== 'http') {
                this.$router.push({
                    path: '/admin/server/edit',
                    query: {name: server.proxyPass, protocol: server.protocol}
                })
            } else {
                this.$router.push({
                    path: '/admin/server/edit',
                    query: {name: server.domains[0], protocol: server.protocol}
                })
            }
        },
        deleteServer(queries) {
            let self = this;
            let url = "/admin/api/directive";
            for (let i = 0; i < queries.length; i++) {
                if (i === 0) {
                    url += "?q=" + encodeURI(queries[i]);
                } else {
                    url += "&q=" + encodeURI(queries[i]);
                }
            }
            this.$axios.delete(url).then(res => {
                self.$toast.success("删除成功！")
                self.queryServices()
            }).catch(e => {
                self.$toast.error(e.message);
            })
        }
    }
}
</script>
