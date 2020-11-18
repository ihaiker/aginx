<template>
    <div>
        <v-title title="负载均衡管理" title-class="icons cui-puzzle">
            <router-link to="/admin/upstream/edit" class="btn btn-sm btn-primary">
                添加负载
            </router-link>
        </v-title>
        <div class="p-3">
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
                <template v-for="up in upstreams">
                    <tr>
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
            </table>
        </div>
    </div>
</template>

<script>
import VTitle from "@/plugins/vTitle";
import Delete from "@/plugins/delete";

export default {
    name: "Upstreams",
    components: {Delete, VTitle},
    data: () => ({
        upstreams: [],
    }),
    mounted() {
        this.queryUpstreams();
    },
    methods: {
        queryUpstreams() {
            let self = this;
            self.$axios.get("/admin/api/upstream").then(res => {
                self.upstreams = res;
            }).catch(e => {
                self.$alert("错误" + e.message);
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
    }
}
</script>
