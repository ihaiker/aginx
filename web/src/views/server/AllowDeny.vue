<template>
    <div class="form-inline">
        <div class="form-group mb-0">
            <div class="input-group">
                <div class="input-group-prepend">
                    <span class="input-group-text">Allow</span>
                </div>
                <template v-for="(allow,idx) in data.allows">
                    <input class="form-control" v-model="data.allows[idx]" type="text">
                    <div class="input-group-append">
                        <button class="btn btn-danger" @click="data.allows.splice(idx,1)">
                            <i class="fa fa-trash"></i>
                        </button>
                    </div>
                </template>
                <div class="input-group-append">
                    <button class="btn btn-css3" @click="appendOne('allows',localIP)">
                        <i class="fa fa-plus"></i>
                    </button>
                </div>
                <div class="input-group-prepend">
                    <span class="input-group-text">Deny</span>
                </div>
                <template v-for="(deny,idx) in data.denys">
                    <input class="form-control" width="120px" v-model="data.denys[idx]" type="text">
                    <div class="input-group-append">
                        <button class="btn btn-danger" @click="data.denys.splice(idx,1)">
                            <i class="fa fa-trash"></i>
                        </button>
                    </div>
                </template>
                <div class="input-group-append">
                    <button class="btn btn-css3" @click="appendOne('denys','all')">
                        <i class="fa fa-plus"></i>
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
export default {
    name: "AllowDeny",
    model: {prop: 'data', event: 'change'},
    props: ['data'],
    data: () => ({
        localIP: "",
    }),
    mounted() {
        this.queryLocal();
    },
    methods: {
        queryLocal() {
            /*let self = this;
            self.$axios.get("http://ip-api.com/json").then(res => {
                self.localIP = res.query;
            }).catch(e => {

            })*/
        },
        appendOne(name, value) {
            if (this.data[name] === undefined || this.data[name] === null) {
                this.$set(this.data, name, [])
            }
            this.data[name].push(value);
        },
    }
}
</script>
