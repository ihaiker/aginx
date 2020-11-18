<template>
    <div class="card card-accent-primary" v-if="ssl">
        <div class="card-header">
            SSL信息
        </div>
        <div class="card-body">
            <div class="row">
                <div class="col-auto">
                    <button class="btn btn-default" @click="server.ssl.httpRedirect = !server.ssl.httpRedirect">
                        HTTP 301跳转HTTPS:
                        <i class="fa"
                           :class="server.ssl.httpRedirect?'fa-check-circle text-success':'fa-square-o'"></i>
                    </button>
                </div>
                <div class="col-auto">
                    <SSLProtocols v-model="server.ssl.protocols"/>
                </div>
            </div>

            <div class="form-group mb-0">
                <div class="input-group">
                    <div class="input-group-prepend">
                        <span class="input-group-text">证书文件(crt)：</span>
                    </div>
                    <input v-model="this.server.ssl.certificate"
                           class="form-control" placeholder="证书文件（PEM）" type="text" readonly="readonly">
                    <div class="input-group-prepend">
                        <span class="input-group-text">私钥文件(key)：</span>
                    </div>
                    <input v-model="this.server.ssl.certificateKey"
                           class="form-control" placeholder="证书文件（key）" type="text" readonly="readonly">
                    <div class="input-group-append">
                        <button class="btn btn-css3" @click="queryCerts">
                            <i class="fa fa-folder-open"></i> 选择
                        </button>
                    </div>
                </div>
            </div>
            <Modal :show="certs !== null" title="选择证书" @ok="certs = null" @cancel="certs = null">
                <div style="max-height: 400px" class="overflow-auto">
                    <table class="table table-hover table-bordered mb-0">
                        <thead>
                        <tr>
                            <th>域名</th>
                            <th>过期时间</th>
                            <th>操作</th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr v-for="(cert,idx) in certs">
                            <td>{{ cert.domain }}</td>
                            <td>{{ expireTime(cert.expireTime) }}</td>
                            <td>
                                <button @click="useCert(cert)"
                                        class="btn btn-sm btn-outline-dark">选择
                                </button>
                            </td>
                        </tr>
                        </tbody>
                    </table>
                </div>
            </Modal>
        </div>
    </div>
</template>

<script>
import SSLProtocols from "@/views/server/SSLProtocols";
import Modal from "@/plugins/modal";

export default {
    name: "ServerSSL",
    components: {Modal, SSLProtocols},
    model: {
        prop: 'server', event: 'change',
    },
    props: ['server'],
    data: () => ({
        certs: null,
    }),
    computed: {
        ssl() {
            if (this.server.listens) {
                for (let i = 0; i < this.server.listens.length; i++) {
                    if (this.server.listens[i].ssl) {
                        return true;
                    }
                }
            }
            return false
        },
    },
    watch: {
        ssl(newVal, oldVal) {
            if (newVal) {
                this.$set(this.server, 'ssl', {
                    httpRedirect: false, protocols: 'TLSv1 TLSv1.1 TLSv1.2',
                    certificate: '', certificateKey: '',
                });
            } else {
                this.$delete(this.server, 'ssl');
            }
        }
    },
    methods: {
        expireTime(t) {
            return t.substr(0, 10) + " " + t.substr(11, 8)
        },
        queryCerts() {
            let self = this;
            self.$axios.get("/admin/api/cert/list").then(res => {
                self.certs = res;
            }).catch(e => {
                self.$alert("查询证书异常：" + e.message);
            });
        },
        useCert(cert) {
            this.server.ssl.certificate = cert.certificate;
            this.server.ssl.certificateKey = cert.privateKey;
            this.certs = null
        }
    }
}
</script>

<style scoped>

</style>
