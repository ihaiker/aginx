<template>
    <div>
        <v-title title="代理列表" title-class="icons cui-puzzle">
            <li class="breadcrumb-menu d-md-down-none" v-if="edit === null">
                <div class="btn-group" role="group" aria-label="Button group">
                    <a class="btn text-danger font-weight-bold" href="javascript:void(0)" @click="addServer">
                        <i class="icon-plus"></i>&nbsp;添加代理</a>
                </div>
            </li>
        </v-title>
        <div v-if="edit === null" class="p-3">
            <table class="table table-bordered table-hover table-striped">
                <thead>
                <tr>
                    <td>域名</td>
                    <td>监听</td>
                    <td>代理地址</td>
                    <td>所在文件</td>
                    <td style="width: 120px;">操作</td>
                </tr>
                </thead>
                <tbody>
                <tr v-for="server in services">
                    <td>
                        <span v-for="(n,idx) in server.name">{{n}}
                            <template v-if="idx !== server.name.length-1">, </template>
                        </span>
                    </td>
                    <td>
                        <span class="bg-success rounded mr-1 p-1" v-for="ls in server.listen">
                            <template v-for="(l,idx) in ls">{{l}}<template v-if="idx !== ls.length-1">&nbsp;</template></template>
                        </span>
                    </td>
                    <td>
                        <div v-for="loc in server.locations">
                            <span class="bg-info rounded mr-1 p-1">
                                <template v-for="(path,idx) in loc.paths">
                                    {{path}}<template v-if="idx !== loc.paths.length-1">&nbsp;</template>
                                </template>
                            </span>
                            <span v-if="loc.type === 'empty'"><span class="badge badge-warning">empty</span></span>
                            <span v-else-if="loc.type === 'root'">{{loc.root}}</span>
                            <span v-else-if="loc.type === 'proxy'">{{loc.loadBalance.proxyType}}://{{loc.loadBalance.proxyAddress}}</span>
                            <span v-else>{{loc.loadBalance.proxyType}}://{{loc.loadBalance.upstream.name}}</span>
                        </div>
                    </td>
                    <td>{{server.from}}</td>
                    <td>
                        <button class="btn btn-default btn-sm" @click="onClickEdit(server)">编辑</button>
                    </td>
                </tr>
                </tbody>
            </table>
        </div>
        <div v-else class="p-3">
            <div class="card">
                <div class="card-header p-2 bg-dark" v-if="edit.from !== ''">
                    服务所在文件：{{edit.from}}
                </div>

                <div class="card-header bg-primary p-2"> 服务名：</div>
                <div class="card-body p-2">
                    <span class="text-primary font-weight-bold">server_name </span>
                    <input type="text" v-for="(n,i) in edit.name"
                           v-autowidth="{maxWidth: '500px', minWidth: '30px', comfortZone: 3}"
                           @input="emptyRemove(edit.name,i)" placeholder="服务名"
                           v-model="edit.name[i]" class="editor mr-3"/>

                    <i class="fa fa-plus-square" @click="edit.name.push('')"></i>
                </div>
                <div class="card-header bg-primary p-2">
                    监听地址：<i class="fa fa-plus" @click="edit.listen.push([''])"></i>
                </div>
                <div class="card-body p-1">
                    <div v-for="(listens,idx) in edit.listen" class="p-1">
                        <span class="text-primary font-weight-bold">listen </span>
                        <input type="text" v-for="(n,i) in listens"
                               v-autowidth="{maxWidth: '500px', minWidth: '30px', comfortZone: 3}"
                               @input="emptyRemove(edit.listen,idx,i)"
                               v-model="edit.listen[idx][i]" class="editor mr-3 text-center"/>
                        <i class="fa fa-plus-square" @click="edit.listen[idx].push('')"></i>
                    </div>
                </div>

                <template v-if="!simple">
                    <div class="card-header bg-primary border-top p-2">
                        额外参数：<i class="fa fa-plus" @click="edit.attrs.push({name:'',attrs:['']})"></i>
                    </div>
                    <div class="card-body p-1">
                        <div class="p-1" v-for="(attr,idx) in edit.attrs">
                            <input type="text" v-autowidth="{minWidth: '20px'}"
                                   @input="attrRemove(edit.attrs,idx)"
                                   v-model="attr.name" class="editor mr-2 text-primary font-weight-bold"/>
                            <input type="text" v-autowidth="{minWidth: '20px'}"
                                   v-for="(att,i) in attr.attrs" @input="attrRemove(edit.attrs,idx,i)"
                                   v-model="edit.attrs[idx].attrs[i]" class="editor mr-2">
                            <i class="fa fa-plus-square" @click="edit.attrs[idx].attrs.push('')"></i>
                        </div>
                    </div>
                </template>

                <div class="card-header bg-primary border-top p-2">
                    配置路径：<i class="fa fa-plus" @click="addLoc"></i>
                </div>
                <div class="card-body p-2">
                    <div v-for="(loc,idx) in edit.locations">

                        <div class="row">
                            <div class="col-auto">
                                <select v-model="loc.type" @change="locTypeChoose(edit.locations[idx])">
                                    <option value="empty">empty</option>
                                    <option value="root">ROOT模式</option>
                                    <option value="proxy">代理模式</option>
                                    <option value="balance">负载均衡</option>
                                </select>
                                <input v-for="(path,i) in loc.paths" v-model="edit.locations[idx].paths[i]"
                                       v-autowidth="{minWidth: '20px'}"
                                       @input="emptyRemove(edit.locations[idx].paths,i)" placeholder="path"
                                       class="editor text-primary font-weight-bold ml-2">
                                <i class="fa fa-plus-square"
                                   @click="edit.locations[idx].paths.push('')"></i>
                            </div>

                            <div v-if="edit.locations[idx].type === 'empty'"></div>
                            <div v-else-if="edit.locations[idx].type === 'root'" class="col-auto">
                                <span class="text-primary font-weight-bold">root：</span>
                                <input v-model="edit.locations[idx].root"
                                       class="editor mr-1" v-autowidth="{minWidth: '20px'}"
                                       type="text" placeholder=" root "> &nbsp;

                                <span class="text-primary font-weight-bold">index: </span>
                                <input v-for="(index,j) in edit.locations[idx].index"
                                       v-model="edit.locations[idx].index[j]"
                                       @input="emptyRemove(edit.locations[idx].index,j)"
                                       class="editor mr-1" v-autowidth="{minWidth: '20px'}"
                                       type="text" placeholder=" root index ">

                                <i class="fa fa-plus-square"
                                   @click="edit.locations[idx].index.push('')"></i>
                            </div>
                            <div v-else class="col-auto">
                                <select v-model="edit.locations[idx].loadBalance.proxyType">
                                    <option value="http">http://</option>
                                    <option value="https">https://</option>
                                </select>

                                <input v-if="edit.locations[idx].type === 'proxy'" type="text"
                                       v-model="edit.locations[idx].loadBalance.proxyAddress"
                                       v-autowidth="{minWidth: '20px'}" class="editor" placeholder=" 代理地址 ">

                                <input v-if="edit.locations[idx].type === 'balance'" type="text"
                                       v-model="edit.locations[idx].loadBalance.upstream.name"
                                       v-autowidth="{minWidth: '20px'}" class="editor" placeholder=" 负载名称 ">
                            </div>
                            <div class="col">
                                <button class="btn btn-xs btn-danger mr-2 pull-right">
                                    <i class="fa fa-remove" @click="edit.locations.splice(idx,1)"></i>
                                </button>
                            </div>

                        </div>

                        <template v-if="!simple">
                            <div class="mt-1">
                                <b>额外参数：</b><i class="fa fa-plus"
                                               @click="edit.locations[idx].attrs.push({name:'',attrs:[]})"></i>
                            </div>
                            <div class="p-1 pl-3" v-for="(attr,i) in loc.attrs">
                                <input v-model="attr.name" v-autowidth="{minWidth: '20px'}"
                                       @input="attrRemove(edit.locations[idx].attrs,i)"
                                       class="editor text-primary font-weight-bold mr-2">

                                <input v-for="(att,j) in attr.attrs" v-model="attr.attrs[j]"
                                       @input="attrRemove(edit.locations[idx].attrs,i,j)"
                                       class="editor mr-2" v-autowidth="{minWidth: '20px'}">

                                <i class="fa fa-plus-square"
                                   @click="edit.locations[idx].attrs[i].attrs.push('')"></i>
                            </div>
                        </template>

                        <template v-if="loc.type === 'balance'">
                            <div class="mt-2">
                                <b>负载地址：</b><span
                                class="text-danger font-weight-bold">{{loc.loadBalance.upstream.name}}</span>
                                <a href="javascript:void(0)" class="ml-3"
                                   @click="edit.locations[idx].loadBalance.upstream.items.push({server:'',attrs:[]})">
                                    <i class="fa fa-plus"></i> 添加负载
                                </a>
                                <a href="javascript:void(0)" class="ml-3" v-if="!simple"
                                   @click="edit.locations[idx].loadBalance.upstream.attrs.push({name:'', attrs:[]})">
                                    <i class="fa fa-plus"></i> 添加负载参数
                                </a>
                            </div>

                            <!-- 路径属性 -->
                            <div v-if="!simple" v-for="(attr,i) in loc.loadBalance.upstream.attrs" class="pl-3">
                                <input type="text" v-model="attr.name"
                                       class="editor text-primary font-weight-bold mr-2"
                                       @input="attrRemove(loc.loadBalance.upstream.attrs,i)"
                                       v-autowidth="{minWidth: '20px'}">

                                <input type="text"
                                       v-for="(at,j) in attr.attrs"
                                       v-model="attr.attrs[j]" @input="emptyRemove(attr.attrs,j)"
                                       class="editor mr-2" v-autowidth="{minWidth: '20px'}">

                                <i class="fa fa-plus-square"
                                   @click="attr.attrs.push('')"></i>
                            </div>

                            <!-- 负载均衡 -->
                            <div v-for="(item,k) in loc.loadBalance.upstream.items" class="pl-3 pt-2">
                                <span class="text-primary font-weight-bold">server </span>
                                <input type="text"
                                       v-model="edit.locations[idx].loadBalance.upstream.items[k].server"
                                       @input="removeUpstreamServer(idx,k)"
                                       class="editor font-weight-bold mr-2" v-autowidth="{minWidth: '20px'}">
                                <input v-for="(att,j) in item.attrs"
                                       v-model="edit.locations[idx].loadBalance.upstream.items[k].attrs[j]"
                                       @input="removeUpstreamServer(idx,k,j)"
                                       class="editor mr-2" v-autowidth="{minWidth: '20px'}">

                                <i class="fa fa-plus-square"
                                   @click="edit.locations[idx].loadBalance.upstream.items[k].attrs.push('')"></i>
                            </div>

                        </template>

                        <hr v-if="idx < edit.locations.length-1"/>
                    </div>
                </div>

                <div class="card-footer">
                    <!-- button class="btn btn-sm mr-3" :class="{'btn-default':simple,'btn-primary':!simple}"
                            @click="simple = !simple">
                        <i v-if="simple" class="fa fa-circle"></i>
                        <i v-else class="fa fa-circle-o"></i>
                        {{simple?'&nbsp;简&nbsp;易&nbsp;模&nbsp;式&nbsp;':'&nbsp;全&nbsp;属&nbsp;性&nbsp;模&nbsp;式&nbsp;'}}
                    </button -->

                    <button class="btn btn-sm btn-outline-primary mr-3" @click="modifyServer">&nbsp;确&nbsp;认&nbsp;</button>
                    <button class="btn btn-sm btn-outline-success mr-3" @click="cancelEdit">&nbsp;取&nbsp;消&nbsp;</button>
                    <Delete v-if="edit.from !== ''" title="" message="确定删除服务" @ok="deleteServer">
                        <button class="btn btn-sm btn-outline-danger mr-3">&nbsp;删&nbsp;除&nbsp;</button>
                    </Delete>
                </div>
            </div>
        </div>
    </div>
</template>
<style>
    .editor {
        border: none;
        border-bottom: 1pt dashed red !important;
    }
</style>
<script>
    import VTitle from "../../plugins/vTitle";
    import VueInputAutowidth from 'vue-input-autowidth'
    import Delete from "../../plugins/delete";

    export default {
        name: "Files",
        components: {Delete, VTitle, VueInputAutowidth},
        data: () => ({
            services: [], edit: null, simple: false,
            editNames: [],
        }),
        mounted() {
            this.queryServices();
        },
        methods: {
            emptyRemove(ary, idx, i) {
                if (i === undefined) {
                    if (ary[idx] === "") {
                        ary.splice(idx, 1)
                    }
                } else {
                    if (ary[idx][i] === "") {
                        ary[idx].splice(i, 1)
                    }
                    if (ary[idx].length === 0) {
                        ary.splice(idx, 1)
                    }
                }
            },
            attrRemove(attrs, idx, i) {
                if (i === undefined) {
                    if (attrs[idx].name === "") {
                        attrs.splice(idx, 1)
                    }
                } else if (attrs[idx].attrs[i] === '') {
                    attrs[idx].attrs.splice(i, 1)
                }
            },
            queryServices() {
                let self = this;
                this.$axios.get("/server").then(res => {
                    self.services = res;
                }).catch(e => {
                    self.$toast.error(e.message);
                })
            },
            locTypeChoose(loc) {
                switch (loc.type) {
                    case "root":
                        loc.root = "";
                        loc.index = [];
                        break
                    case "proxy":
                        loc.loadBalance = {
                            attrs: [], proxyType: 'http',
                            proxyAddress: '127.0.0.1:8080',
                        }
                        break
                    case "balance":
                        let upstreamName = "";
                        if (this.edit.name && this.edit.name.length > 0 && this.edit.name[0] !== '') {
                            upstreamName = this.edit.name[0].replace(new RegExp('\\.', "gm"), "_");
                        }
                        loc.loadBalance = {
                            attrs: [], proxyType: 'http',
                            upstream: {name: upstreamName, attrs: [], items: []}
                        }
                        break;
                }
            },
            addLoc() {
                if (!this.edit.locations) {
                    this.edit.locations = [];
                }
                this.edit.locations.push({
                    type: 'proxy',
                    paths: ['/'],
                    attrs: [],
                    root: '', index: [],
                    loadBalance: {
                        proxyType: 'http', proxyAddress: '127.0.0.1:8080', attrs: [],
                        upstream: {name: '', attrs: [], items: []}
                    }
                })
            },
            removeUpstreamServer(idx, k, j) {
                if (this.edit.locations[idx].loadBalance.upstream.items[k].server === "") {
                    this.edit.locations[idx].loadBalance.upstream.items.splice(k, 1);
                }
                if (j !== undefined && this.edit.locations[idx].loadBalance.upstream.items[k].attrs[j] === '') {
                    this.edit.locations[idx].loadBalance.upstream.items[k].attrs.splice(j, 1);
                }
            },
            deleteServer() {
                let self = this;
                self.$axios.delete("/server?q=" + self.editNames[0]).then(res => {
                    self.edit = null;
                    self.queryServices();
                }).catch(e => {
                    self.$toast.error(e.message);
                })
            },
            modifyServer() {
                let self = this;
                let url = "/server?" + this.param({"q": this.editNames});
                self.$axios.post(url, self.edit).then(res => {
                    self.edit = null;
                    self.queryServices();
                }).catch(e => {
                    self.$toast.success(e.message);
                })
            },
            onClickEdit(edit) {
                this.edit = edit;
                this.editNames = [];
                for (let i = 0; i < this.edit.name.length; i++) {
                    this.editNames.push(this.edit.name[i]);
                }
            },
            cancelEdit() {
                this.edit = null;
                this.queryServices();
            },
            param(params) {
                return Object.keys(params).map(function (k) {
                    return encodeURIComponent(k) + '=' + encodeURIComponent(params[k])
                }).join('&');
            },
            addServer() {
                this.edit = {
                    "from": "", "listen": [["80"]], "name": [""],
                    "attrs": [{"name": "try_files", "attrs": ["$uri", "@tornado"]}],
                    "locations": [
                        {
                            "type": "balance", "paths": ["@tornado"],
                            "attrs": [
                                {"name": "proxy_set_header", "attrs": ["X-Scheme", "$scheme"]},
                                {"name": "proxy_set_header", "attrs": ["Host", "$host"]},
                                {"name": "proxy_set_header", "attrs": ["X-Real-IP", "$remote_addr"]},
                                {"name": "proxy_set_header", "attrs": ["X-Forwarded-For", "$proxy_add_x_forwarded_for"]}
                            ],
                            "loadBalance": {
                                "proxyType": "http",
                                "upstream": {
                                    "from": "", "name": "", "attrs": [],
                                    "items": [{"server": "127.0.0.1:8080", "attrs": []}]
                                }
                            }
                        }
                    ]
                };
                this.editNames = [];
            }
        }
    }
</script>
