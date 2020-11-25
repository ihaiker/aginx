import Vuex from 'vuex'

Vue.use(Vuex)

let code = localStorage.getItem("aginx.node.code") || "";
let name = localStorage.getItem("aginx.node.name") || "未选择";
let token = localStorage.getItem("aginx.token") || ""

const store = new Vuex.Store({
    state: {
        node: {code: code, name: name},
        token: token
    },
    mutations: {
        setNode(state, node) {
            state.node.name = node.name;
            state.node.code = node.code;
            localStorage.setItem("aginx.node.code", node.code);
            localStorage.setItem("aginx.node.name", node.name);
        },
        setToken(state, token) {
            state.token = token
            if (token === "") {
                localStorage.removeItem("aginx.token");
            } else {
                localStorage.setItem("aginx.token", token);
            }
        }
    },
    getters: {
        node: state => state.node,
        token: state => state.token,
    }
})

export default store;
