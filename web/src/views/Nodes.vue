<template>
    <div v-if="!include">
        <v-title title="管理节点" title-class="icons cui-puzzle">
            <a href="#" class="text-danger font-weight-bold" @click="onAddNode">
                <i class="fa fa-plus-circle"></i>&nbsp;添加节点
            </a>
        </v-title>

        <div class="row p-5">
            <div class="col-auto" v-for="(node,idx) in nodes">
                <div class="card" :class="{'bg-primary':activeNode(node)}" @click="chooseNode(node)">
                    <div class="card-header">
                        {{ node.name }}&nbsp;&nbsp;(&nbsp;{{ node.code }}&nbsp;)
                        <div class="card-header-actions">
                            <a @click="onAddNode(node)"
                               class="card-header-action btn-setting text-dark" href="#">
                                <i class="icon-settings"></i>
                            </a>
                        </div>
                    </div>
                    <div class="card-body">
                        <h5>用户：{{ node.user }}</h5>
                        <h6>{{ node.address }}</h6>
                    </div>
                </div>
            </div>
        </div>

        <modal title="添加节点" v-if="node !== null" :show="node !== null" @cancel="node = null" @ok="setNode">
            <div class="p-3">
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">名　称:</span>
                        </div>
                        <input type="text" class="form-control" v-model="node.name" placeholder="节点名称描述">
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">地　址:</span>
                        </div>
                        <input type="text" class="form-control" v-model="node.address" placeholder="节点地址">
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">编　码:</span>
                        </div>
                        <input type="text" class="form-control" v-model="node.code" placeholder="由字母、数字、- _ 组成">
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">用　户:</span>
                        </div>
                        <input type="text" class="form-control" v-model="node.user" placeholder="节点BasicAuth用户">
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">密　码:</span>
                        </div>
                        <input type="text" class="form-control" v-model="node.password" placeholder="节点BasicAuth密码">
                    </div>
                </div>
            </div>
        </modal>

    </div>
    <div v-else class="vh-100">
        <v-title title="节点快速选择" title-class="icons cui-puzzle"/>
        <ul class="list-group list-group-flush vh-100" style="overflow-y: auto">
            <li v-for="(node,idx) in nodes"
                :class="{'active':activeNode(node)}" @click="chooseNode(node)"
                class="list-group-item list-group-item-action">
                {{ node.name }}&nbsp;&nbsp;(&nbsp;{{ node.code }}&nbsp;)
            </li>
        </ul>
    </div>
</template>

<script>

import VTitle from "@/plugins/vTitle";
import Modal from "@/plugins/modal";

export default {
    name: "Nodes", components: {Modal, VTitle},
    data: () => ({
        nodes: [], node: null,
        colors: ["text-white bg-success", "bg-info", "bg-warning", "bg-danger"]
    }),
    inject: ["reload"],
    props: {
        include: {
            type: Boolean,
            default: false
        }
    },
    mounted() {
        this.queryNodes();
    },
    methods: {
        onAddNode(n) {
            if (n === undefined || n === null) {
                this.node = {
                    code: '', name: '',
                    user: '', password: '', address: '',
                }
            } else {
                this.node = {
                    code: n.code, name: n.name,
                    user: n.user, password: '', address: n.address,
                }
            }
        },
        queryNodes() {
            let self = this;
            self.$axios.get("/admin/nodes").then(res => {
                self.nodes = res;
            }).catch(e => {
                self.$alert(e.message);
            });
        },

        chooseNode(node) {
            this.$store.commit("setNode", node);
            //this.$toast.success("节点切换成功,当前节点：" + node.name);
        },
        setNode() {
            if (this.node.name === "") {
                this.$alert("名称不能为空！");
                return
            }
            if (this.node.code === "") {
                this.$alert("节点编码不能为空！");
                return
            }
            if (this.node.address === "") {
                this.$alert("节点地址不能为空！");
                return
            }
            if (this.node.user === "") {
                this.$alert("节点用户不能为空！");
                return
            }
            if (this.node.password === "") {
                this.$alert("节点密码不能为空！");
                return
            }
            let self = this;
            self.$axios.post("/admin/node", this.node).then(res => {
                self.node = null;
                self.queryNodes();
            }).catch(e => {
                self.$alert('设置错误！');
            })
        },
        activeNode(node) {
            return this.$store.getters.node.code === node.code
        }
    }
}
</script>
<style scoped>
</style>
