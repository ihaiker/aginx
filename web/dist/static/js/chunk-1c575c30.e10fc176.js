(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-1c575c30"],{"0167":function(t,e,i){"use strict";i("ec9c")},3938:function(t,e,i){"use strict";i.r(e);var a=function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("div",[i("v-title",{attrs:{title:"代理列表","title-class":"icons cui-puzzle"}},[i("router-link",{attrs:{to:"/admin/server/edit"}},[i("i",{staticClass:"fa fa-plus-circle"}),t._v(" 添加服务\n        ")])],1),i("div",{staticClass:"p-3"},[i("div",{staticClass:"form-group form-inline"},[i("div",{staticClass:"input-group"},[t._m(0),i("input",{directives:[{name:"model",rawName:"v-model",value:t.searchName,expression:"searchName"}],staticClass:"form-control",attrs:{type:"text",placeholder:"域名"},domProps:{value:t.searchName},on:{keyup:function(e){return!e.type.indexOf("key")&&t._k(e.keyCode,"enter",13,e.key,"Enter")?null:t.queryServices(e)},input:function(e){e.target.composing||(t.searchName=e.target.value)}}}),i("div",{staticClass:"input-group-append"},[i("button",{staticClass:"btn btn-primary",on:{click:t.queryServices}},[i("i",{staticClass:"fa fa-search-plus"}),t._v(" 搜索\n                    ")])])])]),i("table",{staticClass:"table table-bordered table-hover"},[t._m(1),i("tbody",[t._l(t.services,(function(e,a){return[t.showPage(a)?i("tr",[i("td",[i("span",{staticClass:"badge badge-dark"},[t._v(t._s(e.protocol))]),t._l(e.domains,(function(e){return i("span",{staticClass:"text-success font-weight-bold ml-2"},[t._v("\n                            "+t._s(e)+"\n                        ")])})),e.commit?i("div",{staticClass:"text-black-50"},[t._v("\n                            "+t._s(e.commit)+"\n                        ")]):t._e()],2),i("td",t._l(e.listens,(function(e){return i("div",[i("span",[t._v(t._s(e.host)+":"+t._s(e.port))]),e.default?i("span",{staticClass:"badge badge-success ml-2"},[t._v("默认")]):t._e(),e.http2?i("span",{staticClass:"badge badge-info ml-2"},[t._v("http2")]):t._e(),e.ssl?i("span",{staticClass:"badge badge-danger ml-2"},[t._v("ssl")]):t._e()])})),0),i("td",["http"!==e.protocol?i("div",[t._v("\n                            转向负载：\n                            "),i("router-link",{staticClass:"text-primary font-weight-bold",attrs:{to:{path:"/admin/upstream/edit",query:{name:e.proxyPass}}}},[t._v("\n                                "+t._s(e.proxyPass)+"\n                            ")])],1):t._e(),t._l(e.locations,(function(a){return"http"===e.protocol?i("div",[i("div",{staticClass:"badge badge-light"},[t._v("\n                                "+t._s(a.path)+"\n                            ")]),"html"===a.type?[t._v("\n                                静态文件\n                                "),i("span",{staticClass:"text-success font-weight-bold"},[t._v("\n                                    "+t._s(a.html.model)+": "+t._s(a.html.path)+"\n                                ")])]:"upstream"===a.type?[t._v("\n                                负载均衡\n                                "),i("router-link",{staticClass:"text-primary font-weight-bold",attrs:{to:{path:"/admin/upstream/edit",query:{name:a.upstream.name}}}},[t._v("\n                                    "+t._s(a.upstream.name)+"\n                                ")]),t._v("\n                                "+t._s(a.upstream.path)+"\n                            ")]:"http"===a.type?[t._v("\n                                动态代理\n                                "),i("span",{staticClass:"text-primary font-weight-bold"},[t._v("\n                                    "+t._s(a.http.to)+"\n                                ")])]:[t._v("\n                                用户定义\n                            ")],a.commit?i("span",{staticClass:"text-secondary"},[t._v(t._s(a.commit))]):t._e()],2):t._e()}))],2),i("td",[i("div",{staticClass:"d-flex justify-content-around"},[i("button",{staticClass:"btn btn-sm btn-outline-primary",on:{click:function(i){return t.editServer(e)}}},[i("i",{staticClass:"fa fa-edit"}),t._v(" 编辑\n                            ")]),i("button",{staticClass:"btn btn-sm btn-outline-danger",on:{click:function(i){return t.deleteServer(e.queries)}}},[i("i",{staticClass:"fa fa-remove"}),t._v(" 删除\n                            ")])])])]):t._e()]}))],2),i("tfoot",[i("tr",[i("td",{attrs:{colspan:"4"}},[i("x-page",{attrs:{items:t.page},on:{change:function(e){t.page.page=e}}})],1)])])])])],1)},s=[function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("div",{staticClass:"input-group-prepend"},[i("span",{staticClass:"input-group-text"},[t._v("域名")])])},function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("thead",[i("tr",[i("th",{staticClass:"text-wrap",staticStyle:{width:"240px"},attrs:{scope:"col"}},[t._v("协议/域名/描述")]),i("th",{staticStyle:{width:"100px"},attrs:{scope:"col"}},[t._v("监听")]),i("th",{attrs:{scope:"col"}},[t._v("代理地址")]),i("th",{staticStyle:{width:"160px"},attrs:{scope:"col"}},[t._v("操作")])])])}],n=i("c0cf"),r=i("f174"),o=i.n(r),l=i("c9b1"),c=i("3dcf"),u={name:"Files",components:{XPage:c["a"],Delete:l["a"],VTitle:n["a"],VueInputAutowidth:o.a},data:function(){return{services:[],searchName:"",page:{page:1,total:0,limit:12}}},mounted:function(){this.queryServices()},methods:{showPage:function(t){return t>=(this.page.page-1)*this.page.limit&&t<this.page.page*this.page.limit},refresh:function(){this.queryServices()},queryServices:function(){this.startLoading(),this.page.page=1;var t=this,e="/admin/api/server";""!==this.searchName&&(e+="?name="+encodeURI(this.searchName)),this.$axios.get(e).then((function(e){t.services=e,t.page.total=t.services.length})).catch((function(e){t.$toast.error(e.message)})).finally((function(){t.finishLoading()}))},editServer:function(t){"http"!==t.protocol?this.$router.push({path:"/admin/server/edit",query:{name:t.proxyPass,protocol:t.protocol}}):this.$router.push({path:"/admin/server/edit",query:{name:t.domains[0],protocol:t.protocol}})},deleteServer:function(t){for(var e=this,i="/admin/api/directive",a=0;a<t.length;a++)i+=0===a?"?q="+encodeURI(t[a]):"&q="+encodeURI(t[a]);this.$axios.delete(i).then((function(t){e.$toast.success("删除成功！"),e.queryServices()})).catch((function(t){e.$toast.error(t.message)}))}}},d=u,p=(i("0167"),i("2877")),m=Object(p["a"])(d,a,s,!1,null,null,null);e["default"]=m.exports},"3dcf":function(t,e,i){"use strict";var a=function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("nav",{attrs:{"aria-label":"navigation"}},[i("ul",{staticClass:"pagination pagination-sm justify-content-center mb-0"},[i("li",{staticClass:"page-item disabled"},[i("a",{staticClass:"page-link",attrs:{href:"#",tabindex:"-1","aria-disabled":"true"}},[t._v("\n                共 "+t._s(t.total)+" 条\n            ")])]),i("li",{staticClass:"page-item",class:{disabled:t.start}},[i("button",{staticClass:"page-link",on:{click:function(e){return t.toPage(1)}}},[t._v("首页")])]),t._l(t.pages,(function(e){return i("li",{staticClass:"page-item",class:{active:e===t.page}},[i("button",{staticClass:"page-link",on:{click:function(i){return t.toPage(e)}}},[t._v(t._s(e))])])})),i("li",{staticClass:"page-item",class:{disabled:t.end}},[i("button",{staticClass:"page-link",on:{click:function(e){return t.toPage(t.pages.length)}}},[t._v("尾页")])]),i("li",{staticClass:"page-item disabled"},[i("a",{staticClass:"page-link",attrs:{href:"#",tabindex:"-1","aria-disabled":"true"}},[t._v("\n                每页 "+t._s(t.limit)+" 条\n            ")])])],2)])},s=[],n={name:"XPage",props:{items:Object},methods:{toPage:function(t){this.$emit("change",t)}},computed:{start:function(){return 1===this.page||this.total<this.limit},end:function(){return this.page===this.pages.length||this.total<this.limit},total:function(){return this.items.total},page:function(){return this.items.page},limit:function(){return this.items.limit},pages:function(){var t=[];console.log("total:",this.total,", limit:",this.limit," pages: ",Math.ceil(this.total/this.limit));for(var e=1;e<=Math.ceil(this.total/this.limit);e++)t.push(e);return t}}},r=n,o=i("2877"),l=Object(o["a"])(r,a,s,!1,null,null,null);e["a"]=l.exports},"9c4a":function(t,e,i){"use strict";i("9eff").polyfill()},"9eff":function(t,e,i){"use strict";function a(t,e){if(void 0===t||null===t)throw new TypeError("Cannot convert first argument to object");for(var i=Object(t),a=1;a<arguments.length;a++){var s=arguments[a];if(void 0!==s&&null!==s)for(var n=Object.keys(Object(s)),r=0,o=n.length;r<o;r++){var l=n[r],c=Object.getOwnPropertyDescriptor(s,l);void 0!==c&&c.enumerable&&(i[l]=s[l])}}return i}function s(){Object.assign||Object.defineProperty(Object,"assign",{enumerable:!1,configurable:!0,writable:!0,value:a})}t.exports={assign:a,polyfill:s}},c0cf:function(t,e,i){"use strict";var a=function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("ol",{staticClass:"breadcrumb breadcrumb-fixed"},[i("li",{staticClass:"breadcrumb-item"},[i("i",{staticClass:"fa",class:t.titleClass}),t._v(" "+t._s(t.title)+"\n    ")]),i("li",{staticClass:"ml-auto"},[t._t("default")],2)])},s=[],n={name:"vTitle",props:{title:String,titleClass:{type:String,default:""}}},r=n,o=i("2877"),l=Object(o["a"])(r,a,s,!1,null,null,null);e["a"]=l.exports},c9b1:function(t,e,i){"use strict";var a=function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("span",{on:{click:t.confirm}},[t._t("default")],2)},s=[],n={name:"delete",props:{title:{type:String,default:"确定？"},message:{type:String,default:"确定删除"}},methods:{confirm:function(){var t=this;this.$confirm(t.message,{title:t.title}).then((function(e){t.$emit("ok")})).catch((function(t){}))}}},r=n,o=i("2877"),l=Object(o["a"])(r,a,s,!1,null,null,null);e["a"]=l.exports},ec9c:function(t,e,i){},f174:function(t,e,i){"use strict";function a(t,e){var i=document.querySelector(".vue-input-autowidth-mirror-".concat(t.dataset.uuid)),a={maxWidth:"none",minWidth:"none",comfortZone:0},s=Object.assign({},a,e.value);t.style.maxWidth=s.maxWidth,t.style.minWidth=s.minWidth;var n=t.value;n||(n=t.placeholder||"");while(i.childNodes.length)i.removeChild(i.childNodes[0]);i.appendChild(document.createTextNode(n));var r=i.scrollWidth+s.comfortZone+2;r!=t.scrollWidth&&(t.style.width="".concat(r,"px"))}i("9c4a");var s={bind:function(t){if("INPUT"!==t.tagName.toLocaleUpperCase())throw new Error("v-input-autowidth can only be used on input elements.");t.dataset.uuid=Math.random().toString(36).slice(-5),t.style.boxSizing="content-box"},inserted:function(t,e){var i=window.getComputedStyle(t);t.mirror=document.createElement("span"),Object.assign(t.mirror.style,{position:"absolute",top:"0",left:"0",visibility:"hidden",height:"0",overflow:"hidden",whiteSpace:"pre",fontSize:i.fontSize,fontFamily:i.fontFamily,fontWeight:i.fontWeight,fontStyle:i.fontStyle,letterSpacing:i.letterSpacing,textTransform:i.textTransform}),t.mirror.classList.add("vue-input-autowidth-mirror-".concat(t.dataset.uuid)),t.mirror.setAttribute("aria-hidden","true"),document.body.appendChild(t.mirror),a(t,e)},componentUpdated:function(t,e){a(t,e)},unbind:function(t){document.body.removeChild(t.mirror)}},n=function(t){t.directive("autowidth",s)};"undefined"!==typeof window&&window.Vue&&window.Vue.use(n),s.install=n,t.exports=s}}]);
//# sourceMappingURL=chunk-1c575c30.e10fc176.js.map