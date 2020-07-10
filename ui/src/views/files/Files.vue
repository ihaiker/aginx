<template>
    <div>
        <ol class="breadcrumb">
            <li class="breadcrumb-fixed">
                <a href="javascript:void(0)" class="text-dark" @click="upFolder(-1)"><i class="fa fa-home"/>根目录</a>：
            </li>
            <li v-for="(p,idx) in prefix" class="breadcrumb-fixed text-primary font-weight-bold">
                <a href="javascript:void(0)" @click="upFolder(idx)">/{{p}}</a>
            </li>
            <li v-if="checkName === null" class="breadcrumb-fixed text-danger pl-3">
                <button class="btn btn-xs btn-outline-danger" @click="setCheck('','')">
                    <i class="fa fa-plus-circle"/> 添加文件
                </button>
                <button class="btn btn-xs btn-outline-dark" @click="upFolder()" :disabled="prefix.length === 0">
                    <i class="fa fa-chevron-circle-up"/> 上层文件夹
                </button>
            </li>
        </ol>

        <div v-if="show" class="p-3">
            <div class="form-group">
                <div class="row">
                    <div class="col-9">
                        <div class="input-group">
                            <div class="input-group-prepend"><span class="input-group-text">文件名称</span></div>
                            <div v-if="prefix.join('/') !== ''" class="input-group-prepend">
                                <span class="input-group-text">{{prefix.join("/")}}</span>
                            </div>
                            <input class="form-control" v-model="checkName" type="text" name="program"
                                   placeholder="文件名称"/>
                        </div>
                    </div>
                    <div class="col-auto">
                        <button class="btn btn-linkedin" @click="setCheck(null,'')">&nbsp;取&nbsp;消&nbsp;</button>
                        <button class="btn btn-primary ml-2" :disabled="checkName === ''" @click="modifyFiles">&nbsp;更&nbsp;新&nbsp;</button>
                        <Delete v-if="checkName !== ''" :message="'您确定要删除文件：' + checkName" @ok="removeFiles(checkName)">
                            <button class="btn btn-danger ml-2">&nbsp;删&nbsp;除&nbsp;</button>
                        </Delete>
                    </div>
                </div>
            </div>
            <codemirror v-model="checkVal" :options="cmOptions"></codemirror>
        </div>

        <div v-else class="pl-5 pt-3">
            <div class="row">
                <div class="col-auto" v-for="item in showFiles">
                    <div v-if="item.folder" class="brand-card-body" @click="setFolder(item.path)">
                        <div class="p-1">
                            <i class="fa fa-2x text-warning fa-folder"/>
                            <div style="max-width: 120px;" class="text-wrap">{{item.name}}</div>
                        </div>
                    </div>
                    <div v-else class="brand-card-body" @click="setCheck(item.name,item.value)">
                        <div class="p-1">
                            <i class="fa fa-2x text-info fa-file"/>
                            <div style="max-width: 120px;" class="text-wrap">{{item.name}}</div>
                        </div>
                    </div>
                </div>
                <!-- div class="col-auto" @click="setCheck('','')">
                    <div class="border p-3 pl-4 pr-4 rounded">
                        <i class="fa fa-2x fa-plus-circle text-danger"></i>
                        <div class="text-nowrap text-danger">添加</div>
                    </div>
                </div>
                <div v-if="prefix.length !== 0" class="col-auto cursor-move" @click="upFolder">
                    <div class="border pl-4 pr-4 rounded">
                        <div class="">
                            <i class="fa fa-chevron-circle-up fa-2x text-danger"></i>
                            <div class="text-nowrap text-danger">..</div>
                        </div>
                    </div>
                </div -->
            </div>
        </div>
    </div>
</template>

<style>
    .CodeMirror {
        height: 50%;
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

    export default {
        name: "Files",
        components: {Modal, Delete, VTitle, codemirror},
        data: () => ({
            prefix: [],
            files: [], checkName: null, checkVal: "",
            cmOptions: {
                tabSize: 4,
                theme: 'lesser-dark', mode: 'nginx',
                line: true, lineWrapping: true, lineNumbers: true,
                collapseIdentical: false, highlightDifferences: true
            }
        }),
        mounted() {
            this.queryFiles();
        },
        computed: {
            show() {
                return this.checkName !== null;
            },
            showFiles() {
                let pathPrefix = this.prefix.join("/")
                let folders = {};
                let folderFiles = [];

                for (let name in this.files) {
                    let paths = name.split("/")
                    let f = {
                        name: paths.pop(), path: name,
                        folder: false, value: this.files[name],
                    }
                    if (pathPrefix === paths.join("/")) {
                        folderFiles.push(f)
                    }
                    if (paths.length > 0) {
                        for (let i = 0; i < 100; i++) {
                            let folderName = paths[paths.length - 1]
                            let folderPath = paths.join("/")
                            if (paths.pop() === undefined) {
                                break
                            }
                            if (paths.join("/") === pathPrefix) {
                                folders[folderPath] = {
                                    name: folderName, path: folderPath, folder: true, value: "",
                                }
                            }
                        }
                    }
                }

                let shows = [];
                for (let i in folders) {
                    shows.push(folders[i])
                }
                for (let i in folderFiles) {
                    shows.push(folderFiles[i])
                }
                return shows
            }
        },
        methods: {
            setCheck(name, val) {
                this.checkName = name;
                this.checkVal = val;
            },
            setFolder(path) {
                this.prefix = path.split("/")
            },
            upFolder(i) {
                if (i === undefined) {
                    this.prefix.pop();
                } else {
                    this.prefix = this.prefix.slice(0, i + 1);
                }
            },
            queryFiles() {
                let self = this;
                self.$axios.get("/file").then(res => {
                    self.files = res;
                }).catch(e => {
                    self.$alert(e.message);
                });
            },
            modifyFiles() {
                let self = this;
                self.$axios.post("/file/ctx", {
                    file: self.getFilePath(this.checkName),
                    body: this.checkVal
                }).then(res => {
                    self.$toast.success("更新成功！");
                    self.setCheck(null, '');
                    self.queryFiles();
                }).catch(e => {
                    self.$alert(e.message);
                });
            },
            removeFiles(name) {
                let self = this;
                self.$axios.delete("/file?file=" + self.getFilePath(name)).then(res => {
                    self.setCheck(null, '');
                    self.queryFiles();
                }).catch(e => {
                    self.$alert(e.message);
                });
            },
            getFilePath(name) {
                let filePath = this.prefix.join("/")
                if (filePath !== "") {
                    filePath += "/"
                }
                filePath += name
                return filePath
            }
        }
    }
</script>
