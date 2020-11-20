<template>
    <div>
        <v-title title="Basic认证管理" title-class="icons cui-puzzle">
            <button class="btn btn-outline-dark" @click="addAuthFile">
                <i class="fa fa-user-circle-o"></i> 添加密码文件
            </button>
        </v-title>

        <div class="p-3">
            <div class="row">
                <div v-for="authFile in authFiles" class="col-lg-4 col-md-6 col-sm-12">
                    <table class="table table-bordered table-hover">
                        <tr>
                            <td colspan="2">
                                <div class="pull-right">
                                    <button @click="editFile = authFile; editUser.name = ''; editUser.passwd = ''"
                                            class="btn btn-sm btn-outline-dark">
                                        <i class="fa fa-user-plus"></i> 添加用户
                                    </button>
                                    <Delete :message="'确定删除文件：'+authFile.name" @click="removeFile(authFile.name)">
                                        <button class="btn btn-sm btn-outline-danger ml-1">
                                            <i class="fa fa-trash"></i> 删除文件
                                        </button>
                                    </Delete>
                                </div>
                                文件：{{ authFile.name }}
                            </td>
                        </tr>
                        <tr>
                            <th>用户</th>
                            <th>操作</th>
                        </tr>
                        <tr v-for="user in authFile.users">
                            <td>{{ user.name }}</td>
                            <td>
                                <Delete @click="removeUser(authFile, user.name)" :message="'您确定删除用户'+user.name">
                                    <button class="btn btn-sm btn-outline-danger">
                                        <i class="fa fa-remove"></i> 删除
                                    </button>
                                </Delete>
                                <button @click="editFile = authFile; editUser.name = user.name; editUser.passwd = ''"
                                        class="btn btn-sm btn-outline-dark ml-2">
                                    <i class="icon-settings"></i> 设置密码
                                </button>
                            </td>
                        </tr>
                    </table>
                </div>
            </div>
        </div>

        <modal title="添加用户" :show="editFile !== null" @cancel="editFile = null" @ok="setAuthUser">
            <div class="p-3">
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">用户：</span>
                        </div>
                        <input v-model="editUser.name" class="form-control"
                               :class="{'is-invalid':editUser.name === ''}" type="text">
                    </div>
                </div>
                <div class="form-group mb-0">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">密码：</span>
                        </div>
                        <input v-model="editUser.passwd" class="form-control"
                               :class="{'is-invalid':editUser.passwd === ''}" type="text">
                        <div class="input-group-append">
                            <button class="btn btn-default" @click="randomPassword">
                                <i class="fa fa-random"></i>随机
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </modal>
    </div>
</template>

<script>
import apr1 from '../tools/apr1'
import VTitle from "@/plugins/vTitle";
import {Base64} from 'js-base64'
import Modal from "@/plugins/modal";
import Delete from "@/plugins/delete";

export default {
    name: "BasicAuths",
    components: {Delete, Modal, VTitle},
    mounted() {
        this.refresh();
    },
    data: () => ({
        editFile: null,
        editUser: {name: '', passwd: ''},
        authFiles: [],
    }),
    methods: {
        refresh() {
            this.queryBasicAuthFile();
        },
        convertAuths(auths) {
            this.authFiles = [];
            for (let i = 0; i < auths.length; i++) {
                let authFile = auths[i];
                this.authFiles.push({
                    name: authFile.name, users: []
                })
                let users = Base64.decode(authFile.content).split("\n");
                users.forEach((u) => {
                    if (u === "" || u.startsWith("#")) {
                        return
                    }
                    let uap = u.split(":", 2)
                    this.authFiles[i].users.push({name: uap[0], passwd: uap[1]});
                })
            }
        },

        queryBasicAuthFile() {
            this.startLoading();
            let self = this;
            let url = "/admin/api/file/search?q=" + encodeURI("auths/*")
            self.$axios.get(url).then(this.convertAuths).catch(e => {
                self.$alert("查询失败：" + e.message);
            }).finally(() => {
                self.finishLoading();
            })
        },

        addAuthFile() {
            let fileName = window.prompt("文件名称");
            if (fileName === null) {

            } else if (fileName.trim() === "") {
                this.$alert("文件名不能为空！")
            } else {
                this.setAuthFile("auths/" + fileName, "\n");
            }
        },

        setAuthUser() {
            if (this.editUser.name === '' || this.editUser.passwd === '') {
                this.$alert("用户名或者密码不能为空")
            } else {
                let name = this.editFile.name
                let content = this.editUser.name + ":" + apr1.hash(this.editUser.passwd) + "\n";
                for (let i = 0; i < this.editFile.users.length; i++) {
                    let user = this.editFile.users[i];
                    if (user.name !== this.editUser.name) {
                        content += user.name + ":" + user.passwd + "\n";
                    }
                }
                this.setAuthFile(name, content)
                this.editFile = null;
            }
        },
        removeUser(file, name) {
            let content = "";
            for (let i = 0; i < file.users.length; i++) {
                let user = file.users[i];
                if (user.name !== name) {
                    content += user.name + ":" + user.passwd + "\n";
                }
            }
            content += '\n'
            this.setAuthFile(file.name, content)
        },
        removeFile(fileName) {
            let self = this;
            self.$axios.delete("/admin/api/file?q=" + encodeURI(fileName)).then(res => {
                self.$toast.success("删除成功！");
                self.queryBasicAuthFile();
            }).catch(e => {
                self.$alert(e.message);
            });
        },
        setAuthFile(name, content) {
            let self = this;
            let formData = new FormData();
            formData.append('path', name);
            formData.append('fileContext', content);
            self.$axios.post("/admin/api/file", formData, {
                headers: {'Content-Type': 'multipart/form-data'}
            }).then(res => {
                self.$toast.success("更新成功！");
                self.queryBasicAuthFile();
            }).catch(e => {
                self.$alert(e.message);
            });
        },

        randomPassword() {
            let len = 16;
            var chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';
            var maxPos = chars.length;
            var pwd = '';
            for (var i = 0; i < len; i++) {
                pwd += chars.charAt(Math.floor(Math.random() * maxPos));
            }
            this.editUser.passwd = pwd;
        }
    },
}
</script>
