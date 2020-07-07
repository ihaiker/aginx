<template>
    <div>
        <v-title title="文件管理" title-class="fa fa-user"/>
        <div class="animated fadeIn p-3">

        </div>
    </div>
</template>

<script>
    import VTitle from "../../plugins/vTitle";

    export default {
        name: "Files",
        components: {VTitle},
        data: () => ({}),
        mounted() {
            this.queryFiles();
        },
        methods: {
            queryFiles() {
                let self = this;
                self.$axios.get("/file", {}).then(res => {
                    console.log(res);
                }).catch(e => {
                    self.$alert(e.message);
                });
            }
        }
    }
</script>
