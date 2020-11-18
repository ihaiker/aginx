import Vuex from 'vuex'

Vue.use(Vuex)

let code = localStorage.getItem("aginx.node.code") || "";
let name = localStorage.getItem("aginx.node.name") || "未选择";

const store = new Vuex.Store({
    state: {
        node: {code: code, name: name}
    },
    mutations: {
        setNode(state, node) {
            state.node.name = node.name;
            state.node.code = node.code;
            localStorage.setItem("aginx.node.code", node.code);
            localStorage.setItem("aginx.node.name", node.name);
        },
    },
    getters: {
        node: state => state.node,
    }
})

export default store;
