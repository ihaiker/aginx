<template>
    <!-- basic auth -->
    <div class="form-group">
        <div class="input-group">
            <div class="input-group-prepend">
                <button class="btn btn-default" @click="basicAuthSwitch">
                    开启BasicAuth:
                    <i class="fa" :class="isBasicAuth?'fa-check-circle text-success':'fa-square-o'"></i>
                </button>
            </div>
            <template v-if="isBasicAuth">
                <div class="input-group-prepend">
                    <span class="input-group-text">认证文件</span>
                </div>
                <select v-model="data.authBasic.userFile" class="form-control">
                    <option v-for="f in authFiles" :value="f.name">{{ f.name }}</option>
                </select>
            </template>
        </div>
    </div>
</template>

<script>
export default {
    name: "BasicAuth",
    model: {prop: 'data', event: 'change'},
    props: ['data'],
    data: () => ({
        authFiles: [],
    }),
    mounted() {
        this.queryBasicAuthFile();
    },
    computed: {
        isBasicAuth() {
            if (this.data.authBasic) {
                return this.data.authBasic.switch === 'on'
            }
            return false
        },
    },
    methods: {
        queryBasicAuthFile() {
            let self = this;
            self.$axios.get("/admin/api/file/search?q=" + encodeURI("auths/*"))
                .then(res => {
                    self.authFiles = res;
                }).catch(e => {
                self.$alert("查询失败：" + e.message);
            })
        },
        basicAuthSwitch() {
            if (this.isBasicAuth) {
                this.$delete(this.data, 'authBasic');
            } else {
                this.$set(this.data, 'authBasic', {
                    switch: 'on', userFile: ''
                });
            }
        }
    }
}
</script>

<style scoped>

</style>
