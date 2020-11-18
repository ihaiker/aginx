import axios from 'axios'
import main from '../main'

axios.defaults.baseURL = process.env.VUE_APP_URL;
axios.defaults.timeout = 15000;
axios.defaults.withCredentials = true;
axios.defaults.headers.post['Content-Type'] = 'application/json;charset=UTF-8';

// 添加请求拦截器
axios.interceptors.request.use(function (config) {
    config.headers['Aginxnode'] = main.$store.getters.node.code;
    return config
});

//添加响应拦截器
axios.interceptors.response.use((response) => {
    if (response.status === 200 && response.data) {
        return response.data
    }
    return response
}, function (err) {
    if (err.response && err.response.data) {
        if (err.response.data.message.indexOf("未发现节点") != -1) {
            main.$toast.error("节点未发现：请先选择节点");
            main.$router.push({path: '/signin', replace: true});
            return
        }
        return Promise.reject(err.response.data)
    } else {
        return Promise.reject({e: err, message: err.message})
    }
});

let config = {
    transformRequest: [function (data) {
        let ret = '';
        for (let it in data) {
            ret += encodeURIComponent(it) + '=' + encodeURIComponent(data[it]) + '&'
        }
        return ret
    }],
    headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
    }
};

export default {
    axios: axios,
    form: config
}
