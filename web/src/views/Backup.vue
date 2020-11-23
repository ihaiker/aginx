<template>
    <div>
        <v-title title="备份管理" title-class="icons cui-puzzle"/>
        <div class="p-3 row">
            <div class="col-4">
                <div class="card">
                    <div class="card-header">
                        备份列表
                    </div>
                    <ul class="list-group">
                        <li v-for="(b,idx) in backups" class="list-group-item">
                            <div class="pull-right">
                                <button class="btn btn-sm btn-primary" @click="rollback(b)">
                                    <i class="icons cui-action-undo"></i> 恢复
                                </button>
                                <button class="btn btn-sm btn-danger ml-2" @click="deleteBackup(b.name)">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
                            <span class="text-primary font-weight-bold">{{ b.name }}</span>
                            <p class="text-black-50">{{ b.comment }}</p>
                        </li>

                        <li class="list-group-item list-group-item-action">
                            <button class="btn btn-primary btn-block" @click="name = ''">
                                备份
                            </button>
                        </li>
                    </ul>
                </div>
            </div>
        </div>
        <modal title="备份文件" :show="name !== null" @cancel="name = null" @ok="backup">
            <div class="p-3">
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">备份文件备注：</span>
                        </div>
                        <input class="form-control" type="text" v-model="name" placeholder="备份文件备注"/>
                    </div>
                </div>
            </div>
        </modal>
    </div>
</template>

<script>
import VTitle from "@/plugins/vTitle";
import Modal from "@/plugins/modal";

export default {
    name: "Backup",
    components: {Modal, VTitle},
    data: () => ({
        backups: [],
        name: null,
    }),
    mounted() {
        this.queryBackups()
    },
    methods: {
        queryBackups() {
            let self = this;
            self.$axios.get("/admin/api/backup").then(res => {
                self.backups = res;
            }).catch(e => {
                self.$alert(e.message);
            })
        },
        backup() {
            let self = this;
            self.startLoading("正在备份")
            self.$axios.post("/admin/api/backup?comment=" + encodeURI(self.name)).then(res => {
                self.$toast.success("备份成功: " + res.name)
                self.queryBackups();
            }).catch(e => {
                self.$alert(e.message);
            }).finally(() => {
                self.finishLoading();
            })
        },
        deleteBackup(name) {
            let self = this;
            self.startLoading("删除备份")
            self.$axios.delete("/admin/api/backup?name=" + name).then(res => {
                self.$toast.success("删除成功: ")
                self.queryBackups();
            }).catch(e => {
                self.$alert(e.message);
            }).finally(() => {
                self.finishLoading();
            })
        },
        rollback(name) {
            let self = this;
            self.startLoading("恢复中。。")
            self.$axios.put("/admin/api/backup?name=" + name).then(res => {
                self.$toast.success("恢复成功: " + res)
                self.queryBackups();
            }).catch(e => {
                self.$alert(e.message);
            }).finally(() => {
                self.finishLoading();
            })
        }
    }
}
</script>

<style scoped>


</style>
