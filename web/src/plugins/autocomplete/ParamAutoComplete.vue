<template>
    <VueBootstrapTypehead :input-class="inputClass"
                          :serializer="serializer"
                          :value="value" min-matching-chars="0" :data="data"
                          @hit="valueChange" @input="inputChange"
                          @keyup="watchRemove"/>
</template>

<script>
import VueBootstrapTypehead from "@/plugins/autocomplete/VueBootstrapTypeahead";

export default {
    name: 'ParamAutoComplete',
    components: {VueBootstrapTypehead},
    model: {
        prop: 'value', event: 'change'
    },
    props: {
        value: {
            type: String,
            default: '',
        },
        prompt: String,
        inputClass: {
            type: String,
            default: '',
        },
        data: {
            type: Array,
            required: true,
            validator: d => d instanceof Array
        },
    },
    data: () => ({
        remove: false,
    }),
    methods: {
        valueChange(v) {
            this.value = v.name;
            this.$emit('change', v.name)
            if (v.args) {
                this.$emit("args", v.args)
            }
        },
        inputChange(v) {
            this.value = v;
            this.$emit('change', v)
        },
        watchRemove(v) {
            if (v.key === 'Backspace' && this.value === '') {
                if (this.remove) {
                    this.$emit("remove", true);
                } else {
                    this.remove = true;
                    let self = this;
                    setTimeout(() => {
                        self.remove = false;
                    }, 500);
                }
            } else {
                this.remove = false;
            }
        },
        serializer: function (p) {
            let txt = '<p class="mb-0 font-weight-bold">' + p.name + '</p>';
            if (p.desc !== '') {
                txt += '<small>' + p.desc + '</small>';
            }
            return txt;
        }
    }
}
</script>
