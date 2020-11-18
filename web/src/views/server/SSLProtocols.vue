<template>
    <div class="form-group">
        <div class="input-group">
            <div class="input-group-prepend">
                <span class="input-group-text">HTTPS支持协议</span>
            </div>

            <div v-for="p in protocols" class="input-group-append">
                <button class="btn btn-default" @click="setSupport(p)">
                    <i class="fa" :class="isSupport(p)?'fa-check-circle text-success':'fa-square-o'"/>
                    {{ p }}
                </button>
            </div>

        </div>
    </div>
</template>

<script>
export default {
    name: "SSLProtocols",
    data: () => ({
        protocols: ["SSLv2", "SSLv3", "TLSv1", "TLSv1.1", "TLSv1.2", "TLSv1.3"]
    }),
    model: {
        prop: 'value', event: 'change'
    },
    props: ['value'],
    computed: {
        supports() {
            return this.value.split(" ")
        }
    },
    methods: {
        isSupport(protocol) {
            for (let i = 0; i < this.supports.length; i++) {
                if (this.supports[i] === protocol) {
                    return true
                }
            }
            return false
        },
        setSupport(protocol) {
            if (this.isSupport(protocol)) {
                let newValue = ''
                for (let i = 0; i < this.supports.length; i++) {
                    if (protocol !== this.supports[i]) {
                        newValue += this.supports[i] + ' ';
                    }
                }
                this.value = newValue.trim();
            } else {
                this.value = this.value + (this.value.trim() === '' ? '' : ' ') + protocol;
            }
            this.$emit('change', this.value);
        }
    }
}
</script>
