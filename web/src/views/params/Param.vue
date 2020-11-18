<template>
    <div>
        <v-title title="参数管理" title-class="icons cui-list"></v-title>
        <div class="animated fadeIn p-3">
            <div class="btn-group" role="group" aria-label="Basic example">
                <button class="btn" @click="setTab('')"
                        :class="tab === ''?'btn-primary':'btn-default'" type="button">
                    &nbsp;&nbsp;基本参数&nbsp;&nbsp;
                </button>
                <button class="btn" @click="setTab('http')"
                        :class="tab === 'http'?'btn-primary':'btn-default'" type="button">
                    &nbsp;&nbsp;HTTP参数&nbsp;&nbsp;
                </button>
                <button class="btn" @click="setTab('stream')"
                        :class="tab === 'stream'?'btn-primary':'btn-default'" type="button">
                    &nbsp;&nbsp;TCP参数&nbsp;&nbsp;
                </button>
            </div>

            <table class="table table-bordered table-hover table-sm mt-3">
                <thead class="thead-dark">
                <tr>
                    <th style="width: 200px">参数名</th>
                    <th>值</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="(c) in config"
                    v-if="c.name !== 'stream' && c.name !== 'http' && c.name !== 'server' && c.name !== 'upstream' ">
                    <td>{{ c.name }}</td>
                    <td>
                        <div class="row">
                            <div class="col-auto" v-for="arg in c.args">{{ arg }}</div>
                        </div>
                        <table v-if="c.name !== 'include' && c.body !== undefined && c.body.length > 0"
                               class="table table-bordered table-sm mb-0 pt-3">
                            <tbody>
                            <tr v-for="b in c.body">
                                <td style="width: 200px;">{{ b.name }}</td>
                                <td><span class="mr-2" v-for="arg in b.args">{{ arg }}</span></td>
                            </tr>
                            </tbody>
                        </table>
                    </td>
                </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
<style scoped>
.btn-group button.btn.btn-css3:before {
    font-family: simple-line-icons;
    content: "\E080";
}
</style>
<script>
import VTitle from "@/plugins/vTitle";

export default {
    name: "Param",
    components: {VTitle},
    data: () => ({
        tab: "", config: [],
    }),
    mounted() {
        this.getConfig();
    },
    methods: {
        setTab(name) {
            this.tab = name;
            this.getConfig();
        },
        getConfig() {
            let self = this;
            let url = "/admin/api/directive"
            if (this.tab !== "") {
                url += "?q=" + this.tab + "&q=*";
            }
            self.$axios.get(url).then(res => {
                self.config = res;
            }).catch(e => {
                self.$toast.error(e.message);
            });
        },
    }
}
</script>
