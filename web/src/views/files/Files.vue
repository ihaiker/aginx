<template>
    <div>
        <ol class="breadcrumb">
            <li class="breadcrumb-fixed">
                <router-link class="text-primary font-weight-bold" to="/admin/files"><i class="fa fa-home"/>文件目录：
                </router-link>
            </li>
            <li v-for="(p,idx) in paths" class="breadcrumb-fixed text-primary font-weight-bold">
                <router-link :to="{path:'/admin/files',query:{path:getPath(idx)}}">{{ p }}/</router-link>
            </li>
        </ol>

        <div class="animated fadeIn pl-3 pr-3 pt-3 row">
            <div class="form-group col-12">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text bg-css3 text-white">文件查找</span>
                    </div>
                    <input type="text" name="program" v-model="search" placeholder="匹配内容：*、*.conf、hosts.d/*.conf"
                           class="form-control"/>
                    <div class="input-group-append">
                        <button class="btn btn-sm btn-css3" @click="queryFiles">
                            <i class="fa fa-search"></i>&nbsp;&nbsp;搜&nbsp;&nbsp;索&nbsp;&nbsp;
                        </button>
                        <button class="btn btn-sm btn-primary text-white font-weight-bold"
                                @click="$router.push({path:'/admin/file/edit',query:{path:folder}})">
                            <i class="fa fa-file-text"></i>&nbsp;新建文件&nbsp;
                        </button>
                    </div>
                </div>
            </div>
        </div>
        <div class="pl-4 pr-4">
            <div v-if="search === ''">
                <!--<div class="row">
                    <div v-if="paths.length > 0" class="col-auto">
                        <div class="brand-card-body">
                            <router-link class="text-dark" to="/admin/files">
                                <i class="fa fa-2x text-dark fa-home"/>
                                <div class="text-wrap">根目录</div>
                            </router-link>
                        </div>
                    </div>

                    <div v-if="paths.length > 1" class="col-auto">
                        <div class="brand-card-body">
                            <router-link class="text-dark"
                                         :to="{path:'/admin/files',query:{path:getPath(paths.length-2)}}">
                                <i class="fa fa-2x text-dark fa-folder-open"/>
                                <div class="text-wrap">上一级</div>
                            </router-link>
                        </div>
                    </div>
                </div>-->
                <div class="row">
                    <div class="col-auto" v-for="item in showFiles">
                        <div v-if="item.folder" class="brand-card-body">
                            <router-link :to="{path:'/admin/files', query:{path:item.path} }" class="p-1">
                                <i class="fa fa-2x text-warning fa-folder"/>
                                <div class="text-wrap">{{ item.name }}</div>
                            </router-link>
                        </div>
                        <div v-else class="brand-card-body">
                            <router-link :to="{path:'/admin/file/edit',query:{name:item.path}}"
                                         class="p-1">
                                <i class="fa fa-2x text-info fa-file"/>
                                <div class="text-wrap">{{ item.name }}</div>
                            </router-link>
                        </div>
                    </div>
                </div>
            </div>
            <div v-else class=" pt-1">
                <div class="list-group">
                    <a v-for="(f,idx) in files" class="list-group-item list-group-item-action">
                        <i class="fa text-warning text-info fa-file"/> {{ f.name }}
                    </a>
                </div>
            </div>
        </div>
    </div>
</template>

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
        files: [], search: ""
    }),
    mounted() {
        this.queryFiles();
    },
    computed: {
        folder() {
            let folder = this.$route.query["path"]
            if (folder === undefined) {
                folder = ""
            }
            return folder;
        },
        paths() {
            let folder = this.folder;
            if (folder !== "") {
                return folder.split("/")
            } else {
                return [];
            }
        },
        showFiles() {
            let folder = this.folder;

            let folders = {};
            let shows = [];
            for (let idx in this.files) {
                let f = this.getFile(this.files[idx])
                if (f.dir === folder) {
                    shows.push(f)
                }
                if (f.dir !== "") {
                    let fd = this.getFolder(f.dir, folder)
                    if (fd !== undefined && folders[fd.name] === undefined) {
                        shows.unshift(fd)
                        folders[fd.name] = fd
                    }
                }
            }
            return shows
        }
    },
    methods: {
        refresh() {
            this.queryFiles();
        },
        queryFiles() {
            this.startLoading();
            let self = this;
            let url = "/admin/api/file/search"
            if (this.search !== "") {
                url += "?q=" + encodeURI(this.search);
            }
            self.$axios.get(url).then(res => {
                self.files = res;
            }).catch(e => {
                self.$toast.error(e.message);
            }).finally(() => {
                self.finishLoading();
            });
        },
        getFile(file) {
            let filePaths = file.name.split("/")
            return {
                name: filePaths.pop(), dir: filePaths.join("/"),
                path: file.name, folder: false
            }
        },
        getFolder(file, folder) {
            if (folder === "") {
                let name = file.split("/").shift()
                return {
                    name: name, path: name, folder: true,
                }
            } else if (file.indexOf(folder + "/") === 0 && file !== folder) {
                let name = file.replace(folder + "/", "").split("/").shift();
                return {
                    name: name, path: folder + "/" + name, folder: true,
                }
            }
            return undefined
        },
        getPath(idx) {
            return this.paths.slice(0, idx + 1).join("/")
        },
    }
}
</script>
