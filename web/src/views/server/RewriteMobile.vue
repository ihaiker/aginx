<template>
    <div>
        <div class="form-group">
            <div class="input-group">
                <div class="input-group-prepend">
                    <button class="btn btn-default" @click="switchSupport">
                        开始自动转向手机端:
                        <i class="fa" :class="support?'fa-check-circle text-success':'fa-square-o'"></i>
                    </button>
                </div>
                <div class="input-group-append">
                    <div class="input-group-text">
                        开启后手机访问网页将自动转向手机网页
                    </div>
                </div>
                <template v-if="support">
                    <div class="input-group-prepend">
                        <span class="input-group-text btn-default">转向地址：</span>
                    </div>
                    <input class="form-control" v-model="server.rewriteMobile.domain" placeholder="转向地址" type="text">
                </template>
            </div>
        </div>
        <div v-if="support" class="form-group">
            <label class="form-col-form-label" for="inputSuccess1">手机端匹配agents</label>
            <textarea v-model="server.rewriteMobile.agents" class="form-control is-valid" style="height: 140px">
            </textarea>
        </div>
    </div>
</template>

<script>
export default {
    name: "RewriteMobile",
    model: {
        prop: 'server',
        event: 'change',
    },
    props: ['server'],
    computed: {
        support() {
            return this.server.rewriteMobile !== undefined;
        }
    },
    methods: {
        switchSupport() {
            if (this.support) {
                this.$delete(this.server, "rewriteMobile")
            } else {
                this.$set(this.server, "rewriteMobile", {
                    agents: '"(MIDP)|(WAP)|(UP.Browser)|(Smartphone)|(Obigo)|(Mobile)|(AU.Browser)|(wxd.Mms)|' +
                        '(WxdB.Browser)|(CLDC)|(UP.Link)|(KM.Browser)|(UCWEB)|(SEMC-Browser)|' +
                        '(Mini)|(Symbian)|(Palm)|(Nokia)|(Panasonic)|(MOT-)|(SonyEricsson)|' +
                        '(NEC-)|(Alcatel)|(Ericsson)|(BENQ)|(BenQ)|(Amoisonic)|(Amoi-)|' +
                        '(Capitel)|(PHILIPS)|(SAMSUNG)|(Lenovo)|(Mitsu)|(Motorola)|(SHARP)|' +
                        '(WAPPER)|(LG-)|(LG/)|(EG900)|(CECT)|(Compal)|(kejian)|(Bird)|(BIRD)|(G900/V1.0)|' +
                        '(Arima)|(CTL)|(TDG)|(Daxian)|(DAXIAN)|(DBTEL)|(Eastcom)|(EASTCOM)|(PANTECH)|' +
                        '(Dopod)|(Haier)|(HAIER)|(KONKA)|(KEJIAN)|(LENOVO)|(Soutec)|(SOUTEC)|(SAGEM)|' +
                        '(SEC-)|(SED-)|(EMOL-)|(INNO55)|(ZTE)|(iPhone)|(Android)|(Windows CE)|(Wget)|' +
                        '(Java)|(curl)|(Opera)"', domain: '',
                })
            }
        }
    }
}
</script>
