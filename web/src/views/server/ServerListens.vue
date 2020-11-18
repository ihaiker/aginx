<template>
    <div class="form-group form-inline">
        <div class="input-group">
            <div class="input-group-prepend">
                <span class="input-group-text">监听地址</span>
            </div>
            <template v-for="(l,idx) in server.listens">
                <input class="form-control" v-model="server.listens[idx].host" type="text"
                       placeholder="IP地址">
                <input class="form-control" :value="server.listens[idx].port"
                       @change="server.listens[idx].port = Number($event.target.value)"
                       type="number" placeholder="PORT">

                <template v-if="server.protocol === 'http'">
                    <div class="input-group-prepend">
                        <button class="btn btn-default" @click="l.default = !l.default">
                            <i class="fa"
                               :class="l.default?'fa-check-square text-success':'fa-square-o'"></i>
                            默认
                        </button>
                    </div>
                    <div class="input-group-prepend">
                        <button class="btn btn-default" @click="l.http2 = !l.http2">
                            <i class="fa" :class="l.http2?'fa-check-square text-success':'fa-square-o'"></i>
                            HTTP2
                        </button>
                    </div>
                    <div class="input-group-prepend">
                        <button class="btn btn-default" @click="l.ssl = !l.ssl">
                            <i class="fa" :class="l.ssl?'fa-check-square text-success':'fa-square-o'"></i>
                            SSL
                        </button>
                    </div>
                </template>

                <div v-if="server.listens.length > 1" class="input-group-append">
                    <button class="btn btn-danger" @click="server.listens.splice(idx,1)">
                        <i class="fa fa-trash"></i>
                    </button>
                </div>
            </template>

            <div class="input-group-append"
                 @click="server.listens.push({host:'', port: 8080,default:false,http2:false,ssl:false})">
                <button class="btn btn-css3">
                    <i class="fa fa-plus"></i>
                </button>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: "ServerListens",
    model: {prop: 'server', event: 'change'},
    props: ['server'],
}
</script>
