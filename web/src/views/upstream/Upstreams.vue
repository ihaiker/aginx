<template>
    <div>
        <v-title title="负载均衡管理" title-class="icons cui-puzzle">
            <router-link to="/admin/upstream/edit" class="btn btn-sm btn-primary">
                添加负载
            </router-link>
        </v-title>
        <div class="p-3">
            <div class="form-group form-inline">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text">名字：</span>
                    </div>
                    <input class="form-control" v-model="searchName" type="text" placeholder="名字"
                           @keyup.enter="queryUpstreams">
                    <div class="input-group-append">
                        <button class="btn btn-primary" @click="queryUpstreams">
                            <i class="fa fa-search-plus"></i> 搜索
                        </button>
                    </div>
                </div>
            </div>

            <table class="table table-bordered table-hover">
                <thead>
                <tr>
                    <th style="width: 200px">名称/备注</th>
                    <th style="width: 120px">转发类型</th>
                    <th style="width: 120px">策略</th>
                    <th>负载</th>
                    <th style="width: 120px">操作</th>
                </tr>
                </thead>
                <tbody>
                <template v-for="(up,idx) in upstreams">
                    <tr v-if="showPage(idx)">
                        <td>
                            {{ up.name }}
                            <span v-if="up.commit !== ''" class="text-black-50"><br/>{{ up.commit }}</span>
                        </td>
                        <td>
                            <span class="badge badge-success">{{ up.protocol }}</span>
                        </td>
                        <td>
                            {{ up.loadStrategy === '' ? '默认' : up.loadStrategy }}
                        </td>
                        <td>
                            <span v-for="(s,idx) in up.servers" class="font-weight-bold">
                                <span v-if="idx !== 0">,&nbsp;&nbsp;</span> {{ s.host }}:{{ s.port }}
                            </span>
                        </td>
                        <td>
                            <button class="btn btn-sm btn-primary" @click="editUpstream(up)">
                                编辑
                            </button>
                            <Delete @ok="deleteUpstream(up)">
                                <button class="btn btn-sm btn-danger ml-2">
                                    删除
                                </button>
                            </Delete>
                        </td>
                    </tr>
                </template>
                </tbody>
                <tfoot>
                <tr>
                    <td colspan="5">
                        <XPage :items="page" @change="page.page = $event"/>
                    </td>
                </tr>
                </tfoot>
            </table>
        </div>
    </div>
</template>

<script>
import VTitle from "@/plugins/vTitle";
import Delete from "@/plugins/delete";
import XPage from "@/plugins/XPage";

export default {
    name: "Upstreams",
    components: {XPage, Delete, VTitle},
    data: () => ({
        upstreams: [], searchName: "",
        page: {
            page: 1, total: 0, limit: 12,
        }
    }),
    mounted() {
        this.queryUpstreams();
    },
    methods: {
        refresh() {
            this.queryUpstreams();
        },
        showPage(idx) {
            return idx >= (this.page.page - 1) * this.page.limit
                && idx < (this.page.page * this.page.limit)
        },
        queryUpstreams() {
            this.startLoading();
            let self = this;
            let url = "/admin/api/upstream";
            if (this.searchName !== "") {
                url += "?name=" + encodeURI(this.searchName);
            }
            self.$axios.get(url).then(res => {
                self.upstreams = res;
                self.page.total = self.upstreams.length;
            }).catch(e => {
                self.$alert("错误" + e.message);
            }).finally(() => {
                self.finishLoading()
            })
        },
        editUpstream(upstream) {
            this.$router.push({
                path: '/admin/upstream/edit',
                query: {name: upstream.name}
            })
        },
        deleteUpstream(upstream) {
            let self = this;
            let url = "/admin/api/directive";
            for (let i = 0; i < upstream.queries.length; i++) {
                if (i === 0) {
                    url += "?q=" + encodeURI(upstream.queries[i]);
                } else {
                    url += "&q=" + encodeURI(upstream.queries[i]);
                }
            }
            this.$axios.delete(url).then(res => {
                self.$toast.success("删除成功！")
                self.queryUpstreams()
            }).catch(e => {
                self.$toast.error(e.message);
            })
        }
    },
}
</script>
