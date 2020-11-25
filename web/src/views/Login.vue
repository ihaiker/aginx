<template>
    <div class="app flex-row align-items-center">
        <div class="container">
            <div class="row justify-content-center">
                <div class="col-md-8">
                    <div class="card-group">
                        <div class="card p-4">
                            <div class="card-body">
                                <h1>用户登录</h1>
                                <p class="text-muted">Sign In to your account</p>
                                <div class="input-group mb-3">
                                    <div class="input-group-prepend">
                                        <span class="input-group-text">
                                          <i class="icon-user"></i>
                                        </span>
                                    </div>
                                    <input v-model="username" class="form-control"
                                           type="text" placeholder="Username">
                                </div>
                                <div class="input-group mb-4">
                                    <div class="input-group-prepend">
                                        <span class="input-group-text">
                                          <i class="icon-lock"></i>
                                        </span>
                                    </div>
                                    <input v-model="password" @keyup.enter="login" class="form-control"
                                           type="password" placeholder="Password">
                                </div>
                                <div class="row justify-content-end">
                                    <div class="col-lg-auto col-sm-12">
                                        <button @click="login" class="btn btn-block btn-primary px-4" type="button">Login</button>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="card text-white bg-primary py-3 d-md-down-none" style="width:44%">
                            <div class="card-body ">
                                <div class="text-center">
                                    <h2>Aginx</h2>
                                    <p class="font-weight-bold">为NGINX添加API/SDK/控制台管理。</p>
                                </div>
                                <div>
                                    <p class="mb-0 border-bottom">※ 支持获取免费ssl证书和配置</p>
                                    <p class="mb-0 border-bottom">※ 分布式下使用第三方存储NGINX配置文件，多NGINX可以统一配置和管理。</p>
                                    <p class="mb-0 border-bottom">※ 自动发布docker、consul服务到nginx配置。</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
export default {
    name: "Login",
    data: () => ({
        username: "", password: "",
    }),
    mounted() {
        if (this.$store.getters.token !== "") {
            this.$router.push({path: '/admin', replace: true});
        }
    },
    methods: {
        login() {
            let self = this;
            self.$axios.post("/login", {
                username: self.username, password: self.password,
            }).then(res => {
                self.$store.commit("setToken", res);
                self.$router.push({path: '/admin', replace: true});
            }).catch(e => {
                self.$alert(e.message);
            })
        }
    }
}
</script>
