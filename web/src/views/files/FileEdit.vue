<template>
    <div>
        <ol class="breadcrumb">
            <li class="breadcrumb-fixed">
                <router-link class="text-primary font-weight-bold" to="/admin/files"><i class="fa fa-home"/>文件目录：
                </router-link>
            </li>
        </ol>
        <div class="animated fadeIn pl-3 pt-3 pr-3 row">
            <div class="form-group col-12">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text bg-css3 text-white font-weight-bold">文件路径：</span>
                    </div>
                    <input :disabled="!canEdit" type="text" v-model="fileName" name="program" placeholder=""
                           class="form-control"/>
                    <div class="input-group-append">
                        <button class="btn btn-css3" @click="setFile">
                            <i class="fa fa-edit"></i>&nbsp;更&nbsp;&nbsp;新&nbsp;
                        </button>
                        <button v-if="!canEdit" @click="removeFile" class="btn btn-danger">
                            <i class="fa fa-remove"></i>&nbsp;删&nbsp;&nbsp;除&nbsp;
                        </button>
                        <button class="btn btn-default" @click="$router.push('/admin/files')">
                            <i class="fa fa-backward"></i>&nbsp;取&nbsp;&nbsp;消&nbsp;
                        </button>
                    </div>
                </div>
            </div>
        </div>
        <div class="pl-3 pr-3">
            <codemirror v-model="fileContext" :options="cmOptions"></codemirror>
        </div>
    </div>
</template>

<style>
.CodeMirror {
    height: 90%;
    min-height: 400px;
    font-size: 14px;
}
</style>

<script>
import VTitle from "../../plugins/vTitle";
import Delete from "../../plugins/delete";
import Modal from "../../plugins/modal";
import {codemirror} from 'vue-codemirror'
import 'codemirror/mode/nginx/nginx.js'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/lesser-dark.css'
import {Base64} from 'js-base64';

export default {
    name: "Files",
    components: {Modal, Delete, VTitle, codemirror},
    data: () => ({
        fileContext: "", fileName: "",
        cmOptions: {
            tabSize: 4, theme: 'lesser-dark', mode: 'nginx',
            line: true, lineWrapping: true, lineNumbers: true,
            collapseIdentical: false, highlightDifferences: true
        }
    }),
    mounted() {
        this.fileName = (this.$route.query["name"] || "");
        if (this.fileName !== "") {
            this.getFile(this.fileName)
        }
        let path = this.$route.query["path"] || ""
        if (path !== "") {
            this.fileName = path + "/" + this.fileName;
        }
    },
    computed: {
        canEdit() {
            let name = this.$route.query["name"]
            return name === undefined || name === ""
        }
    },
    methods: {
        getFile(name) {
            let self = this;
            let url = "/admin/api/file?q=" + encodeURI(name)
            self.$axios.get(url).then(res => {
                self.fileContext = Base64.decode(res.content);
            }).catch(e => {
                self.$toast.error(e.message);
            });
        },
        setFile() {
            let self = this;
            let formData = new FormData();
            formData.append('path', this.fileName);
            formData.append('fileContext', this.fileContext);
            //formData.append("file", this.fileContext, this.fileName)

            self.$axios.post("/admin/api/file", formData, {
                headers: {'Content-Type': 'multipart/form-data'}
            }).then(res => {
                self.$toast.success("更新成功！");
                self.$router.push({path: '/admin/file/edit', query: {name: self.fileName}, replace: true})
            }).catch(e => {
                self.$alert(e.message);
            });
        },
        removeFile() {
            let self = this;
            self.$axios.delete("/admin/api/file?q=" + encodeURI(self.fileName)).then(res => {
                self.$toast.success("删除成功！");
                self.$router.push({path: '/admin/files'})
            }).catch(e => {
                self.$alert(e.message);
            });
        },
    }
}
</script>
