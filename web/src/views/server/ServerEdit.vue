<template>
    <div>
        <v-title title="编辑服务" title-class="icons cui-puzzle">
            <template v-if="server.queries && server.queries.length > 0">
                搜索路径：<span class="badge badge-dark ml-1" v-for="q in server.queries">{{ q }}</span>
            </template>
        </v-title>
        <div class="p-3">
            <!-- server listen -->
            <ServerListens v-model="server"/>

            <!-- domain -->
            <ServerName v-if="server.protocol === 'http'" v-model="server"/>

            <div class="form-group">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text">备注</span>
                    </div>
                    <input class="form-control" v-model="server.commit" type="text">

                    <template v-if="$route.query.name === undefined || server.protocol !== 'http'">
                        <div class="input-group-prepend">
                            <span class="input-group-text">转发类型</span>
                        </div>
                        <select v-model="server.protocol" class="form-control"
                                @change="setServerProtocol($event.target.value)">
                            <option v-if="$route.query.name === undefined"
                                    value="http">HTTP
                            </option>
                            <option value="tcp">TCP</option>
                            <option value="udp">UDP</option>
                        </select>
                    </template>
                </div>
            </div>

            <!-- ssl -->
            <ServerSSL v-if="server.protocol === 'http'" v-model="server"/>

            <ServerLocations v-model="server"/>

            <HideCard title="额外参数">
                <Partamters v-model="server.parameters" prompt="server"/>
                <hr/>
                <!-- basic auth -->
                <BasicAuth v-if="server.protocol === 'http'" v-model="server"/>
                <!-- 手机转向 -->
                <RewriteMobile v-if="server.protocol === 'http'" v-model="server"/>
                <!-- 允许与拒绝 -->
                <AllowDeny v-model="server"/>
            </HideCard>

            <div class="d-flex justify-content-center">
                <button class="btn btn-css3" @click="modfiyServer">
                    确定更新
                </button>
                <router-link to="/admin/servers" class="btn btn-default ml-3" @click="">
                    取　　消
                </router-link>
            </div>
        </div>

       <!-- <codemirror v-model="serverString" :options="cmOptions"></codemirror>-->

    </div>
</template>

<style>
.CodeMirror {
    height: 90%;
    min-height: 400px;
    font-size: 14px;
}

input.form-control {
    color: black;
}

input.form-control::-webkit-input-placeholder {
    color: #cccccc;
}

.form-control[readonly] {
    background-color: transparent;
}
</style>
<script>
import VTitle from "@/plugins/vTitle";
import {codemirror} from 'vue-codemirror'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/lesser-dark.css'
import RewriteMobile from "@/views/server/RewriteMobile";
import BasicAuth from "@/views/server/BasicAuth";
import ServerSSL from "@/views/server/ServerSSL";
import ServerListens from "@/views/server/ServerListens";
import ServerName from "@/views/server/ServerName";
import AllowDeny from "@/views/server/AllowDeny";
import ServerLocations from "@/views/server/ServerLocations";
import HideCard from "@/plugins/HideCard";
import Partamters from "@/views/server/Partamters";

export default {
    name: "ServerEdit",
    components: {
        Partamters, HideCard, ServerLocations,
        AllowDeny, ServerName, ServerListens, ServerSSL, BasicAuth, RewriteMobile, VTitle, codemirror
    },
    data: () => ({
        server: {},
        cmOptions: {
            tabSize: 4, theme: 'lesser-dark', mode: 'json',
            line: true, lineWrapping: true, lineNumbers: true,
            collapseIdentical: false, highlightDifferences: true
        }
    }),
    mounted() {
        if (this.$route.query.name) {
            this.getServer();
        } else {
            this.setServerProtocol('http')
        }

    },
    computed: {
        serverString() {
            return JSON.stringify(this.server, null, "\t")
        },
    },
    methods: {
        setServerProtocol(protocol) {
            if (protocol === 'http') {
                this.server = {
                    "queries": [], "commit": "", "protocol": "http",
                    "listens": [{"port": 80, "default": false, "http2": false, "ssl": false}],
                    "domains": [""],
                    "locations": [{
                        "path": "/", "type": "http", "http": {"to": ""},
                        "basicHeader": true, "webSocket": false, "parameters": []
                    }], "parameters": []
                }
            }
            //编辑tcp的时候不可更改,没有名字说明是编辑非HTTP
            else if (this.$route.query.name === undefined) {
                this.server = {
                    "queries": [], "commit": "", "protocol": protocol,
                    "listens": [{"port": 8080}],
                    "proxyPass": "", "parameters": []
                }
            }
        },
        getServer() {
            let self = this;
            let url = "/admin/api/server?name=" + self.$route.query.name +
                "&protocol=" + self.$route.query.protocol;
            self.$axios.get(url).then(res => {
                self.server = res[0];
            }).catch(e => {
                self.$alert(e.message);
            })
        },
        modfiyServer() {
            let self = this;
            self.$axios.post("/admin/api/server", this.server).then(res => {
                self.$toast.success("更新成功！")
                self.$router.push({path: '/admin/servers'})
            }).catch(e => {
                self.$alert(e.message);
            })
        }
    },
}
</script>
