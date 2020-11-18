<template>
    <div>
        <v-title title="证书管理">
            <button class="btn btn-sm btn-outline-danger"
                    @click="obtainNewDomain={domain:'',provider:'lego'}">申请新证书
            </button>
            <button class="btn btn-sm btn-outline-primary ml-2"
                    @click="custom={domain:''}">添加自定义证书
            </button>
        </v-title>

        <div class="p-3">
            <table class="table table-hover table-bordered">
                <thead>
                <tr>
                    <th>提供商</th>
                    <th>域名</th>
                    <th>证书(crt)</th>
                    <th>证书(key)</th>
                    <th>过期时间</th>
                    <th>操作</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="(cert,idx) in certs">
                    <td>{{ providers[cert.provider] }} ({{ cert.provider }})</td>
                    <td>{{ cert.domain }}</td>
                    <td>{{ cert.certificate }}</td>
                    <td>{{ cert.privateKey }}</td>
                    <td>{{ expireTime(cert.expireTime) }}</td>
                    <td>
                        <button v-if="cert.provider !== 'custom'"
                                @click="obtainDomain(cert.provider,cert.domain)"
                                class="btn btn-sm btn-outline-dark">续租
                        </button>
                        <button v-else @click="custom={domain:cert.domain}"
                                class="btn btn-sm btn-outline-primary">
                            重传
                        </button>
                    </td>
                </tr>
                </tbody>
            </table>
        </div>

        <modal title="申请证书" v-if="obtainNewDomain !== null" :show="obtainNewDomain !== null"
               @cancel="obtainNewDomain = null" @ok="obtainDomainRequest()">
            <div class="p-3">
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">证书供应商：</span>
                        </div>
                        <select v-model="obtainNewDomain.provider" class="form-control">
                            <option v-for="(desc,name) in providers" v-if="name !== 'custom'" :value="name">
                                {{ desc }} ({{ name }})
                            </option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">域名：</span>
                        </div>
                        <input v-model="obtainNewDomain.domain" class="form-control"
                               :class="{'is-invalid':obtainNewDomain.domain === ''}" type="text">
                    </div>
                </div>
            </div>
        </modal>

        <modal title="添加自定义证书" v-if="custom !== null" :show="custom !== null"
               @cancel="custom = null" @ok="uploadCert">
            <div class="p-3">
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">域名：</span>
                        </div>
                        <input v-model="custom.domain" class="form-control" type="text"/>
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">证书（crt）：</span>
                        </div>
                        <input class="form-control" @change="selectFile('crt',$event)" type="file">
                    </div>
                </div>
                <div class="form-group">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text">证书（key）：</span>
                        </div>
                        <input class="form-control" @change="selectFile('key',$event)" type="file">
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
    name: "Certs",
    components: {Modal, VTitle},
    data: () => ({
        certs: [], providers: {},
        obtainNewDomain: null,
        custom: null,
    }),
    mounted() {
        this.queryCerts();
        this.queryInfo();
    },
    methods: {
        queryInfo() {
            let self = this;
            self.$axios.get("/admin/api/info").then(res => {
                self.providers = res.certificate;
            }).catch(e => {
                self.$alert("查询证书异常：" + e.message);
            });
        },
        queryCerts() {
            let self = this;
            self.$axios.get("/admin/api/cert/list").then(res => {
                self.certs = res;
            }).catch(e => {
                self.$alert("查询证书异常：" + e.message);
            });
        },
        expireTime(t) {
            return t.substr(0, 10) + " " + t.substr(11, 8)
        },

        obtainDomainRequest() {
            this.obtainDomain(this.obtainNewDomain.provider, this.obtainNewDomain.domain);
        },

        obtainDomain(provider, domain) {
            this.startLoading("正在申请证书：" + domain);
            let self = this;
            self.$axios.post("/admin/api/cert?domain=" + domain + "&provider=" + provider).then(res => {
                self.$toast.success("申请证书成功！");
                self.queryCerts();
                self.obtainNewDomain = null;
            }).catch(e => {
                self.$alert("申请证书失败：" + e.message);
            }).finally(() => {
                self.finishLoading();
            });
        },

        selectFile(name, f) {
            this.$set(this.custom, name, f);
        },

        uploadCertFile(name, file, cb) {
            let self = this;
            let formData = new FormData();
            formData.append('path', name);
            formData.append("file", file)

            self.$axios.post("/admin/api/file", formData, {
                headers: {'Content-Type': 'multipart/form-data'}
            }).then(res => {
                cb()
            }).catch(e => {
                self.$alert(e.message);
            });
        },

        uploadCert() {
            let i = 0;
            let cb = () => {
                i = i + 1;
                if (i === 2) {
                    this.queryCerts();
                    this.custom = null;
                }
            }
            this.uploadCertFile(
                'certs/custom/' + this.custom.domain + "/server.crt",
                this.custom.crt.target.files[0], cb)

            this.uploadCertFile(
                'certs/custom/' + this.custom.domain + "/server.key",
                this.custom.key.target.files[0], cb)
        }
    }
}
</script>
