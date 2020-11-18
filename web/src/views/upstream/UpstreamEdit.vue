<template>
    <div>
        <v-title title="编辑负载">
            <template v-if="upstream.queries && upstream.queries.length > 0">
                搜索路径：<span class="badge badge-dark ml-1" v-for="q in upstream.queries">{{ q }}</span>
            </template>
        </v-title>
        <div class="p-3">
            <div class="form-group">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text">转发类型</span>
                    </div>
                    <select v-model="upstream.protocol" style="width: 140px" class="form-control">
                        <option value="http">HTTP</option>
                        <option value="tcp">TCP</option>
                        <option value="udp">UDP</option>
                    </select>

                    <div class="input-group-prepend">
                        <span class="input-group-text">名称</span>
                    </div>
                    <input v-model="upstream.name" type="text" class="form-control"/>

                    <div class="input-group-prepend">
                        <span class="input-group-text">负载策略</span>
                    </div>
                    <select v-model="upstream.loadStrategy" class="form-control">
                        <option value="">默认</option>
                        <option value="ip_hash">IP哈希（ip_hash）</option>
                        <option value="fair">页面大小、加载时间长短智能（fair）</option>
                        <option value="url_hash">URL哈希（url_hash）</option>
                        <option value="least_conn">最少连接（least_conn）</option>
                        <option value="least_time">最短时间（least_time）</option>
                        <option value="hash">哈希（hash）</option>
                        <option value="sticky">会话保持 (sticky)</option>
                    </select>
                </div>
            </div>

            <div class="form-group">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text">备注</span>
                    </div>
                    <input v-model="upstream.commit" type="text" class="form-control"/>
                </div>
            </div>

            <table class="table table-hover table-bordered">
                <thead class="thead-light">
                <tr>
                    <td colspan="7">
                        负载
                        <button @click="addServer"
                                class="btn btn-sm btn-outline-danger pull-right">
                            添加负载
                        </button>
                    </td>
                </tr>
                <tr>
                    <td>IP</td>
                    <td>端口</td>
                    <td>权重</td>
                    <td>失败等待时间(s)</td>
                    <td>最大失败次数</td>
                    <td>状态</td>
                    <td width="80px">操作</td>
                </tr>
                </thead>
                <tbody>
                <tr v-for="(server,idx) in upstream.servers">
                    <td>
                        <input v-model="upstream.servers[idx].host"
                               placeholder="IP" type="text" class="form-control"/>
                    </td>
                    <td>
                        <input :value="upstream.servers[idx].port"
                               @change="upstream.servers[idx].port = Number($event.target.value)"
                               placeholder="port" type="number" max="65535" min="1" class="form-control"/>
                    </td>
                    <td>
                        <input :value="upstream.servers[idx].weight"
                               @change="upstream.servers[idx].weight = Number($event.target.value)"
                               placeholder="port" type="number" min="0" class="form-control"/>
                    </td>
                    <td>
                        <input :value="upstream.servers[idx].failTimeout"
                               @change="upstream.servers[idx].failTimeout = Number($event.target.value)"
                               placeholder="port" type="number" min="0" class="form-control"/>
                    </td>
                    <td>
                        <input :value="upstream.servers[idx].maxFails"
                               @change="upstream.servers[idx].maxFails = Number($event.target.value)"
                               placeholder="port" type="number" min="0" class="form-control"/>
                    </td>
                    <td>
                        <select v-model="upstream.servers[idx].status" class="form-control">
                            <option value="">无</option>
                            <option value="down">停用（down）</option>
                            <option value="backup">备用（backup）</option>
                        </select>
                    </td>
                    <td>
                        <button @click="removeServer(idx)"
                                class="btn btn-sm btn-danger">删除
                        </button>
                    </td>
                </tr>
                </tbody>
            </table>

            <HideCard title="额外参数" :show="true">
                <Partamters v-model="upstream.parameters"/>
            </HideCard>
        </div>

        <div class="d-flex justify-content-center">
            <button class="btn btn-primary" @click="modfiyUpstream">　确　定　</button>
            <router-link to="/admin/upstreams" class="btn btn-default ml-3">　取　消　</router-link>
        </div>
    </div>
</template>

<script>
import VTitle from "@/plugins/vTitle";
import Partamters from "@/views/server/Partamters";
import HideCard from "@/plugins/HideCard";
import {codemirror} from 'vue-codemirror'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/lesser-dark.css'

export default {
    name: "UpstreamEdit",
    components: {HideCard, Partamters, VTitle, codemirror},
    data: () => ({
        upstream: {},
        cmOptions: {
            tabSize: 4, theme: 'lesser-dark', mode: 'json',
            line: true, lineWrapping: true, lineNumbers: true,
            collapseIdentical: false, highlightDifferences: true
        }
    }),
    mounted() {
        if (this.$route.query.name) {
            this.getUpstream();
        } else {
            this.upstream = {
                "queries": [], "name": "", "commit": "",
                "protocol": "http",
                "loadStrategy": "",
                "servers": [{
                    "host": "127.0.0.1",
                    "port": 8080,
                    "weight": 0,
                    "failTimeout": 0,
                    "maxFails": 0,
                    "status": ""
                }],
                "parameters": []
            }
        }
    },
    methods: {
        getUpstream() {
            let self = this;
            let url = "/admin/api/upstream?name=" + self.$route.query.name;
            self.$axios.get(url).then(res => {
                self.upstream = res[0];
            }).catch(e => {
                self.$alert(e.message);
            })
        },
        addServer() {
            this.upstream.servers.push({
                "host": "", "port": 8080,
                "weight": 1, "failTimeout": 3, "maxFails": 3,
                "status": ""
            })
        },
        removeServer(idx) {
            this.upstream.servers.splice(idx, 1)
        },
        modfiyUpstream() {
            let self = this;
            self.$axios.post("/admin/api/upstream", this.upstream).then(res => {
                self.$toast.success("更新成功！")
                self.$router.push({path: '/admin/upstreams'})
            }).catch(e => {
                self.$alert(e.message);
            })
        }
    },
    computed: {
        upstreamString() {
            return JSON.stringify(this.upstream, null, "\t")
        },
    },
}
</script>
