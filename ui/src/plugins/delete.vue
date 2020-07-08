<template>
    <span @click="confirm">
        <slot></slot>
    </span>
</template>
<script>
    export default {
        name: "delete",
        props: {
            title: {
                type: String,
                default: "确定？"
            },
            message: {
                type: String,
                default: "确定删除"
            }
        },
        methods: {
            confirm() {
                let self = this;
                this.$confirm(self.message, {"title": self.title}).then(ok => {
                    self.$emit("ok")
                }).catch(cancel => {
                });
            }
        }
    }
</script>
