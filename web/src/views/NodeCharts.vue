<template>
    <div class="card mb-0" v-if="$store.getters.node.code !== ''">
        <div class="card-header bg-white">
            内存/CPU 『{{ $store.getters.node.name }}』
        </div>
        <div class="row">
            <ECharts :options="options" :autoresize="true" class="col-12"/>
        </div>
    </div>
</template>

<script>

import ECharts from 'vue-echarts'
import 'echarts/lib/chart/line'
import 'echarts/lib/component/tooltip'
import 'echarts/lib/component/legend'

const Mem = 0, CPU = 1, NginxMem = 2, NginxCpu = 3;
const legends = ['内存', 'CPU', 'Nginx内存', 'NginxCPU'];

export default {
    name: "NodeCharts",
    components: {ECharts},
    data: () => ({
        maxNum: 30 * 60,
        options: {
            tooltip: {
                trigger: 'axis',
                formatter: function (params) {
                    let tp = "";

                    let finder = function (index) {
                        for (let i = 0; i < params.length; i++) {
                            if (params[i].seriesName === legends[index]) {
                                return params[i].value[1]
                            }
                        }
                        return null;
                    }
                    let mem = finder(0)
                    if (mem !== null) {
                        if (mem > 1024 * 1024) {
                            mem = (mem / 1024 / 1024).toFixed(2) + ' M'
                        } else if (mem > 1024) {
                            mem = (mem / 1024).toFixed(2) + ' K'
                        }
                        tp += '内存：' + mem;
                    }
                    let cpu = finder(1);
                    if (cpu !== null) {
                        tp += '，CPU' + cpu.toFixed(2) + '%';
                    }

                    let nginxMem = finder(2)
                    if (nginxMem !== null) {
                        if (nginxMem > 1024 * 1024) {
                            nginxMem = (nginxMem / 1024 / 1024).toFixed(2) + ' M'
                        } else if (nginxMem > 1024) {
                            nginxMem = (nginxMem / 1024).toFixed(2) + ' K'
                        }
                        tp += '<br/> Nginx 内存：' + nginxMem;
                    }
                    let nginxCpu = finder(3);
                    if (nginxCpu !== null) {
                        tp += '，Nginx CPU: ' + nginxCpu.toFixed(2) + '%';
                    }
                    tp += '<br/>' + params[0].value[0]
                    return tp;
                },
                axisPointer: {
                    animation: false
                }
            },
            xAxis: {
                type: 'time',
                splitLine: {
                    show: false
                },
            },
            yAxis: [{
                type: 'value',
                axisLabel: {
                    formatter(s) {
                        if (s > 1024 * 1024) {
                            return (s / 1024 / 1024).toFixed(2) + ' M'
                        } else if (s > 1024) {
                            return (s / 1024).toFixed(2) + ' K'
                        }
                        return s
                    },
                }
            }, {
                type: 'value',
                splitLine: false,
                boundaryGap: [0, '100%'],
                axisLabel: {
                    formatter(s) {
                        return s.toFixed(1) + "%"
                    }
                }
            }],
            legend: {
                data: legends,
            },
            series: [
                {
                    name: '内存',
                    type: 'line',
                    yAxisIndex: 0,
                    showSymbol: false,
                    hoverAnimation: false,
                    data: []
                },
                {
                    name: 'CPU',
                    type: 'line',
                    showSymbol: false,
                    hoverAnimation: false,
                    yAxisIndex: 1,
                    data: []
                },
                {
                    name: 'Nginx内存',
                    type: 'line',
                    yAxisIndex: 0,
                    showSymbol: false,
                    hoverAnimation: false,
                    data: []
                },
                {
                    name: 'NginxCPU',
                    type: 'line',
                    showSymbol: false,
                    hoverAnimation: false,
                    yAxisIndex: 1,
                    data: []
                }
            ]
        }
    }),
    mounted() {
        let self = this;
        if (self.$store.getters.node.code !== '') {
            self.queryMemInfo(this.maxNum);
        }
        setInterval(() => {
            if (self.$store.getters.node.code !== '') {
                self.queryMemInfo(1)
            }
        }, 1000);
    },
    methods: {
        queryMemInfo(n) {
            let self = this;
            self.$axios.get("/admin/api/memstats?limit=" + n).then(res => {
                if (n === 1) {
                    self.setData([res]);
                } else {
                    self.setData(res);
                }
            }).catch(e => self.$alert(e.message))
        },
        setData(lines) {
            for (let i = 0; i < lines.length; i++) {
                this.options.series[Mem].data.push({
                    name: lines[i].time,
                    value: [lines[i].time, lines[i].mem],
                })
                this.options.series[CPU].data.push({
                    name: lines[i].time,
                    value: [lines[i].time, lines[i].cpu],
                })
                this.options.series[NginxMem].data.push({
                    name: lines[i].time,
                    value: [lines[i].time, lines[i].nginxMem],
                })
                this.options.series[NginxCpu].data.push({
                    name: lines[i].time,
                    value: [lines[i].time, lines[i].nginxCpu],
                })
            }
            if (this.options.series[Mem].data.length > this.maxNum) {
                this.options.series[Mem].data.shift();
                this.options.series[CPU].data.shift();
                this.options.series[NginxMem].data.shift();
                this.options.series[NginxCpu].data.shift();
            }
        }
    },
    watch: {
        selectNode(newNode, oldNode) {
            if (newNode !== oldNode) {
                this.options.series[Mem].data = [];
                this.options.series[CPU].data = [];
                this.options.series[NginxMem].data = [];
                this.options.series[NginxCpu].data = [];
                this.queryMemInfo(this.maxNum);
            }
        }
    },
    computed: {
        selectNode() {
            return this.$store.getters.node.code
        },
    }
}
</script>
