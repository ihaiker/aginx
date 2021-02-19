module.exports = {
    outputDir: "dist", assetsDir: "static",
    lintOnSave: false, runtimeCompiler: true,
    devServer: {
        disableHostCheck: true,
        hot: true, open: true,
    },
    configureWebpack: {
        externals: {
            vue: "Vue",
            'vue-router': 'VueRouter',
            'bootstrap-vue': "BootstrapVue",
        },
    },
    transpileDependencies: [
        'vue-echarts',
        'resize-detector'
    ]
}
