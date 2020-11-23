<template>
    <div v-if="server.protocol === 'http'" class="card card-accent-primary">
        <div class="card-header">
            代理地址(location)
            <div class="card-header-actions">
                <div class="card-header-action btn-minimize text-dark" @click="addLocation">
                    <i class="fa fa-plus"></i>
                    添加代理路径
                </div>
            </div>
        </div>
        <table class="table table-hover mb-0">
            <tr>
                <th style="width: 220px">路径</th>
                <th style="width: 120px">类型</th>
                <th>目标</th>
                <th width="140px">操作</th>
            </tr>
            <template v-if="server.protocol === 'http'" v-for="(loc,idx) in server.locations">
                <tr>
                    <td>
                        <input class="form-control" placeholder="路径"
                               :class="{'is-invalid':server.locations[idx].path===''}"
                               v-model="server.locations[idx].path" type="text">
                    </td>
                    <td>
                        <select v-model="server.locations[idx].type" class="form-control" @change="changeType(idx)">
                            <option value="html">静态文件</option>
                            <option value="upstream">负载均衡</option>
                            <option value="http">动态代理</option>
                            <option value="custom" selected>用户定义</option>
                        </select>
                    </td>
                    <td>
                        <template v-if="server.locations[idx].type === 'html'">
                            <div class="form-group mb-0">
                                <div class="input-group">
                                    <div class="input-group-prepend">
                                        <select v-model="server.locations[idx].html.model" class="form-control">
                                            <option value="root" selected>ROOT模式</option>
                                            <option value="alias">Alias模式</option>
                                        </select>
                                    </div>
                                    <div class="input-group-append">
                                        <span class="input-group-text">路径</span>
                                    </div>
                                    <input class="form-control"
                                           :class="{'is-invalid':server.locations[idx].html.path===''}"
                                           placeholder="路径" v-model="server.locations[idx].html.path" type="text">
                                    <div class="input-group-append">
                                        <span class="input-group-text">主页</span>
                                    </div>
                                    <input class="form-control"
                                           :class="{'is-invalid':server.locations[idx].html.indexes===''}"
                                           placeholder="主页" v-model="server.locations[idx].html.indexes" type="text">
                                </div>
                            </div>
                        </template>
                        <template v-else-if="server.locations[idx].type === 'upstream'">
                            <div class="form-group mb-0">
                                <div class="input-group">
                                    <div class="input-group-prepend">
                                        <span class="input-group-text">负载</span>
                                    </div>
                                    <div class="input-group-append">
                                        <select v-model="server.locations[idx].upstream.name" class="form-control">
                                            <option v-for="upstream in upstreams"
                                                    v-if="upstream.protocol === 'http'" :value="upstream.name">
                                                <template v-if="upstream.commit !== ''">
                                                    {{ upstream.commit }} ({{ upstream.name }})
                                                </template>
                                                <template v-else>
                                                    {{ upstream.name }}
                                                </template>
                                            </option>
                                        </select>
                                    </div>
                                    <div class="input-group-prepend">
                                        <span class="input-group-text">额外地址</span>
                                    </div>
                                    <input class="form-control" v-model="server.locations[idx].upstream.path"
                                           placeholder="额外地址" type="text">

                                    <div class="input-group-prepend">
                                        <button class="btn btn-default"
                                                @click="server.locations[idx].basicHeader = !server.locations[idx].basicHeader">
                                            <i class="fa"
                                               :class="server.locations[idx].basicHeader?'fa-check-square text-success':'fa-square-o'"></i>
                                            基础header
                                        </button>
                                    </div>
                                    <div class="input-group-prepend">
                                        <button class="btn btn-default"
                                                @click="server.locations[idx].webSocket = !server.locations[idx].webSocket">
                                            <i class="fa"
                                               :class="server.locations[idx].webSocket?'fa-check-square text-success':'fa-square-o'"></i>
                                            WebSocket
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </template>
                        <template v-else-if="server.locations[idx].type === 'http'">
                            <div class="form-group mb-0">
                                <div class="input-group">
                                    <div class="input-group-prepend">
                                        <div class="input-group-text">
                                            代理地址：
                                        </div>
                                    </div>
                                    <input class="form-control"
                                           :class="{'is-invalid':server.locations[idx].http.to===''}"
                                           v-model="server.locations[idx].http.to"
                                           placeholder="代理地址" type="text">

                                    <div class="input-group-prepend">
                                        <button class="btn btn-default"
                                                @click="server.locations[idx].basicHeader = !server.locations[idx].basicHeader">
                                            <i class="fa"
                                               :class="server.locations[idx].basicHeader?'fa-check-square text-success':'fa-square-o'"></i>
                                            基础header
                                        </button>
                                    </div>
                                    <div class="input-group-prepend">
                                        <button class="btn btn-default"
                                                @click="server.locations[idx].webSocket = !server.locations[idx].webSocket">
                                            <i class="fa"
                                               :class="server.locations[idx].webSocket?'fa-check-square text-success':'fa-square-o'"></i>
                                            WebSocket
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </template>
                        <template v-else>
                            <!-- custom -->
                        </template>
                    </td>
                    <td class="d-flex justify-content-around">
                        <button class="btn btn-sm btn-css3" title="额外参数" @click="showParamIdx(idx)">
                            <i class="fa fa-align-justify"></i>
                        </button>
                        <button class="btn btn-sm btn-success" @click="moveUp(idx)" :disabled="idx === 0" title="上移动">
                            <i class="fa fa-arrow-up"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" @click="delLoc(idx)" title="删除">
                            <i class="fa fa-trash"></i>
                        </button>
                    </td>
                </tr>
                <tr v-if="showParams === idx">
                    <td colspan="4" class="p-3">
                        <Partamters v-model="server.locations[idx].parameters" prompt="location"/>
                        <hr/>
                        <BasicAuth class="mt-2" v-model="server.locations[idx]"></BasicAuth>
                        <AllowDeny v-model="server.locations[idx]"/>
                    </td>
                </tr>
            </template>
        </table>
    </div>
    <div v-else class="form-group">
        <div class="input-group">
            <div class="input-group-prepend">
                <button class="btn btn-dark ">
                    转向负载：
                </button>
            </div>
            <select v-model="server.proxyPass" class="form-control">
                <option v-for="upstream in upstreams" v-if="upstream.protocol === 'tcp'" :value="upstream.name">
                    <template v-if="upstream.commit !== ''">
                        {{ upstream.commit }} ({{ upstream.name }})
                    </template>
                    <template v-else>
                        {{ upstream.name }}
                    </template>
                </option>
            </select>
        </div>
    </div>
</template>

<script>
import Partamters from "@/views/server/Partamters";
import AllowDeny from "@/views/server/AllowDeny";
import BasicAuth from "@/views/server/BasicAuth";

export default {
    name: "ServerLocations",
    components: {BasicAuth, AllowDeny, Partamters},
    model: {prop: 'server', event: 'change'},
    props: ['server'],
    data: () => ({
        upstreams: [{name: '未找到负载'}],
        showParams: null
    }),
    created() {
        this.queryUpstreams();
    },
    methods: {
        queryUpstreams() {
            let self = this;
            self.$axios.get("/admin/api/upstream").then(res => {
                self.upstreams = res;
            }).catch(e => {
                self.$toast.error("获取负载均衡错误" + e.message);
            })
        },
        changeType(idx) {
            let location = this.server.locations[idx];
            this.$delete(location, "upstream");
            this.$delete(location, "http");
            this.$delete(location, "html");

            if (location.type === 'upstream') {
                this.$set(location, 'upstream', {"name": "", "path": ""})
            } else if (location.type === 'html') {
                this.$set(location, 'html', {
                    "path": "", "model": "root", "indexes": "index.html index.htm"
                })
            } else if (location.type === 'http') {
                this.$set(location, 'http', {"to": ""})
            }
        },
        addLocation() {
            this.server.locations.push({
                path: "", type: 'http',
                http: {to: ""}, parameters: [],
            })
        },
        moveUp(idx) {
            this.server.locations[idx] =
                this.server.locations.splice(idx - 1, 1, this.server.locations[idx])[0];
        },
        delLoc(idx) {
            this.server.locations.splice(idx, 1)
        },
        showParamIdx(idx) {
            if (this.showParams === idx) {
                this.showParams = null;
            } else {
                this.showParams = idx;
            }
        }
    }
}
</script>
