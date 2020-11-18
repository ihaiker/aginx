<template>
    <!-- domain -->
    <div class="form-group form-inline">
        <div class="input-group">
            <div class="input-group-prepend">
                <span class="input-group-text">域名</span>
            </div>
            <template v-for="(domain,idx) in server.domains">
                <input class="form-control" v-model="server.domains[idx]" type="text">
                <div v-if="server.domains.length > 1" class="input-group-append">
                    <button class="btn btn-danger" @click="server.domains.splice(idx,1)">
                        <i class="fa fa-trash"></i>
                    </button>
                </div>
            </template>
            <div class="input-group-append">
                <button class="btn btn-css3" @click="server.domains.push('')">
                    <i class="fa fa-plus"></i>
                </button>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: "ServerName",
    model: {prop: 'server', event: 'change'},
    props: ['server'],
}
</script>

<style scoped>

</style>
