<template>
    <button v-if="parameters && parameters.length === 0" @click="parameters.push({name:'',args:[]})"
            class="btn btn-link text-danger font-weight-bold">
        <i class="fa fa-plus"></i>
        添加参数
    </button>
    <div v-else>
        <template v-for="(attr,idx) in parameters">
            <div class="row">
                <div class="col">
                    <button @click="moveUp(idx)"
                            class="btn btn-outline-success mr-1" title="向上移动">
                        <i class="fa fa-arrow-up"></i>
                    </button>

                    <ParamAutoComplete v-model="parameters[idx].name" :data="namePrompts"
                                       @remove="emptyRemove(idx)" @args="parameters[idx].args = $event"
                                       input-class="text-primary font-weight-bold text-left"/>
                    <ParamAutoComplete v-for="(att,i) in attr.args" :data="argsPrompts"
                                       @remove="attrRemove(idx,i)"
                                       v-model="parameters[idx].args[i]" input-class="ml-2"/>
                </div>

                <div class="col-auto">
                    <button class="btn btn-link" @click="parameters[idx].args.push('')">
                        <i class="fa fa-plus"></i> 属性值
                    </button>

                    <button class="btn btn-link text-dark" @click="addBody(idx)">
                        <i class="fa fa-plus"></i> 子参数
                    </button>

                    <button v-if="idx + 1 === parameters.length" @click="parameters.push({name:'',args:[]})"
                            class="btn btn-link text-danger font-weight-bold">
                        <i class="fa fa-plus"></i>
                        参数
                    </button>
                </div>
            </div>

            <Partamters v-if="parameters[idx].body && parameters[idx].body.length > 0"
                        style="margin-left: 40px;" v-model="parameters[idx].body"/>
        </template>
    </div>
</template>
<script>
import ParamAutoComplete from "@/plugins/autocomplete/ParamAutoComplete";

export default {
    name: "Partamters",
    components: {ParamAutoComplete},
    model: {
        prop: "parameters",
        event: "change"
    },
    props: {
        parameters: {
            type: Array
        },
        prompt: String,
    },
    computed: {
        namePrompts() {
            let ps = this.prompts[this.prompt]
            if (ps === undefined) {
                return []
            }
            return ps.name.concat(this.prompts.all.name);
        },
        argsPrompts() {
            let ps = this.prompts[this.prompt]
            if (ps === undefined) {
                return []
            }
            return ps.args.concat(this.prompts.all.args);
        },
    },
    methods: {
        moveUp(idx) {
            this.parameters[idx] =
                this.parameters.splice(idx - 1, 1, this.parameters[idx])[0];
        },
        addBody(idx) {
            if (this.parameters[idx]['body'] === undefined) {
                this.$set(this.parameters[idx], 'body', []);
            }
            this.parameters[idx].body.push({name: '', args: []})
        },
        emptyRemove(idx) {
            if (this.parameters[idx].name === '') {
                this.parameters.splice(idx, 1);
            }
        },
        attrRemove(idx, i) {
            if (this.parameters[idx].args[i] === '') {
                this.parameters[idx].args.splice(i, 1);
            }
        },
    },
    data: () => ({
        prompts: {
            all: {
                name: [
                    {name: "gzip", desc: "开启GZIP", args: ["on"]},
                    {name: "gzip_buffers", desc: ""},
                    {name: "gzip_comp_level", desc: ""},
                    {name: "gzip_disable", desc: ""},
                    {name: "gzip_http_version", desc: ""},
                    {name: "gzip_min_length", desc: ""},
                    {name: "gzip_proxied", desc: ""},
                    {name: "gzip_types", desc: ""},
                    {name: "gzip_vary", desc: ""},
                    {name: "gunzip", desc: "",},
                    {name: "gunzip_buffers", desc: ""},
                    {name: "gzip_static", desc: ""},
                    {name: "add_header", desc: ""},
                    {name: "add_trailer", desc: ""},
                    {name: "expires", desc: ""},

                    {name: "limit_conn", desc: ""},
                    {name: "limit_conn_dry_run", desc: ""},
                    {name: "limit_conn_log_level", desc: ""},
                    {name: "limit_conn_status", desc: ""},
                    {name: "limit_conn_zone", desc: ""},
                    {name: "limit_zone", desc: ""},

                    {name: "limit_req", desc: ""},
                    {name: "limit_req_dry_run", desc: ""},
                    {name: "limit_req_log_level", desc: ""},
                    {name: "limit_req_status", desc: ""},
                    {name: "limit_req_zone", desc: ""},

                    {name: "access_log", desc: ""},
                    {name: "log_format", desc: ""},
                    {name: "open_log_file_cache", desc: ""},


                    {name: "proxy_bind", desc: ""},
                    {name: "proxy_buffer_size", desc: ""},
                    {name: "proxy_buffering", desc: ""},
                    {name: "proxy_buffers", desc: ""},
                    {name: "proxy_busy_buffers_size", desc: ""},
                    {name: "proxy_cache", desc: ""},
                    {name: "proxy_cache_background_update", desc: ""},
                    {name: "proxy_cache_bypass", desc: ""},
                    {name: "proxy_cache_convert_head", desc: ""},
                    {name: "proxy_cache_key", desc: ""},
                    {name: "proxy_cache_lock", desc: ""},
                    {name: "proxy_cache_lock_age", desc: ""},
                    {name: "proxy_cache_lock_timeout", desc: ""},
                    {name: "proxy_cache_max_range_offset", desc: ""},
                    {name: "proxy_cache_methods", desc: ""},
                    {name: "proxy_cache_min_uses", desc: ""},
                    {name: "proxy_cache_path", desc: ""},
                    {name: "proxy_cache_purge", desc: ""},
                    {name: "proxy_cache_revalidate", desc: ""},
                    {name: "proxy_cache_use_stale", desc: ""},
                    {name: "proxy_cache_valid", desc: ""},
                    {name: "proxy_connect_timeout", desc: ""},
                    {name: "proxy_cookie_domain", desc: ""},
                    {name: "proxy_cookie_flags", desc: ""},
                    {name: "proxy_cookie_path", desc: ""},
                    {name: "proxy_force_ranges", desc: ""},
                    {name: "proxy_headers_hash_bucket_size", desc: ""},
                    {name: "proxy_headers_hash_max_size", desc: ""},
                    {name: "proxy_hide_header", desc: ""},
                    {name: "proxy_http_version", desc: ""},
                    {name: "proxy_ignore_client_abort", desc: ""},
                    {name: "proxy_ignore_headers", desc: ""},
                    {name: "proxy_intercept_errors", desc: ""},
                    {name: "proxy_limit_rate", desc: ""},
                    {name: "proxy_max_temp_file_size", desc: ""},
                    {name: "proxy_method", desc: ""},
                    {name: "proxy_next_upstream", desc: ""},
                    {name: "proxy_next_upstream_timeout", desc: ""},
                    {name: "proxy_next_upstream_tries", desc: ""},
                    {name: "proxy_no_cache", desc: ""},
                    {name: "proxy_pass", desc: ""},
                    {name: "proxy_pass_header", desc: ""},
                    {name: "proxy_pass_request_body", desc: ""},
                    {name: "proxy_pass_request_headers", desc: ""},
                    {name: "proxy_read_timeout", desc: ""},
                    {name: "proxy_redirect", desc: ""},
                    {name: "proxy_request_buffering", desc: ""},
                    {name: "proxy_send_lowat", desc: ""},
                    {name: "proxy_send_timeout", desc: ""},
                    {name: "proxy_set_body", desc: ""},
                    {name: "proxy_set_header", desc: ""},
                    {name: "proxy_socket_keepalive", desc: ""},
                    {name: "proxy_ssl_certificate", desc: ""},
                    {name: "proxy_ssl_certificate_key", desc: ""},
                    {name: "proxy_ssl_ciphers", desc: ""},
                    {name: "proxy_ssl_conf_command", desc: ""},
                    {name: "proxy_ssl_crl", desc: ""},
                    {name: "proxy_ssl_name", desc: ""},
                    {name: "proxy_ssl_password_file", desc: ""},
                    {name: "proxy_ssl_protocols", desc: ""},
                    {name: "proxy_ssl_server_name", desc: ""},
                    {name: "proxy_ssl_session_reuse", desc: ""},
                    {name: "proxy_ssl_trusted_certificate", desc: ""},
                    {name: "proxy_ssl_verify", desc: ""},
                    {name: "proxy_ssl_verify_depth", desc: ""},
                    {name: "proxy_store", desc: ""},
                    {name: "proxy_store_access", desc: ""},
                    {name: "proxy_temp_file_write_size", desc: ""},
                    {name: "proxy_temp_path", desc: ""},
                ],
                args: []
            },
            server: {
                name: [
                    {name: "charset", desc: "字符编码", args: ["'UTF-8'"]},
                    {name: "access_log", desc: "访问日志", args: ['/var/log/nginx/access.log']},
                    {name: "ssl_session_cache", desc: ""},
                    {name: "ssl_session_timeout", desc: "", args: ['60s']},
                    {name: "ssl_ciphers", desc: "SSL支持密码协议"},
                    {name: "ssl_prefer_server_ciphers", desc: "打开SSL密码协议", args: ["on"]},
                    {name: "ssl_buffer_size", desc: ""},
                    {name: "ssl_client_certificate", desc: ""},
                    {name: "ssl_conf_command", desc: ""},
                    {name: "ssl_crl", desc: ""},
                    {name: "ssl_dhparam", desc: ""},
                    {name: "ssl_early_data", desc: ""},
                    {name: "ssl_ecdh_curve", desc: ""},
                    {name: "ssl_ocsp", desc: ""},
                    {name: "ssl_ocsp_cache", desc: ""},
                    {name: "ssl_ocsp_responder", desc: ""},
                    {name: "ssl_password_file", desc: ""},
                    {name: "ssl_prefer_server_ciphers", desc: ""},
                    {name: "ssl_reject_handshake", desc: ""},
                    {name: "ssl_session_cache", desc: ""},
                    {name: "ssl_session_ticket_key", desc: ""},
                    {name: "ssl_session_tickets", desc: ""},
                    {name: "ssl_session_timeout", desc: ""},
                    {name: "ssl_stapling", desc: ""},
                    {name: "ssl_stapling_file", desc: ""},
                    {name: "ssl_stapling_responder", desc: ""},
                    {name: "ssl_stapling_verify", desc: ""},
                    {name: "ssl_trusted_certificate", desc: ""},
                    {name: "ssl_verify_client", desc: ""},
                    {name: "ssl_verify_depth", desc: ""},
                ],
                args: []
            },
            location: {
                name: [
                    {name: "add_before_body", desc: ""},
                    {name: "add_after_body", desc: ""},
                    {name: "auth_request", desc: ""},
                    {name: "auth_request_set", desc: ""},
                    {name: "autoindex", desc: "", args: ["on"]},
                    {name: "autoindex_exact_size", desc: ""},
                    {name: "autoindex_format", desc: ""},
                    {name: "autoindex_localtime", desc: ""},
                    {name: "charset", desc: ""},
                    {name: "charset_map", desc: ""},
                    {name: "charset_types", desc: ""},
                    {name: "override_charset", desc: ""},
                    {name: "source_charset,", desc: ""},
                    {name: "grpc_bind", desc: ""},
                    {name: "grpc_buffer_size", desc: ""},
                    {name: "grpc_connect_timeout", desc: ""},
                    {name: "grpc_hide_header", desc: ""},
                    {name: "grpc_ignore_headers", desc: ""},
                    {name: "grpc_intercept_errors", desc: ""},
                    {name: "grpc_next_upstream", desc: ""},
                    {name: "grpc_next_upstream_timeout", desc: ""},
                    {name: "grpc_next_upstream_tries", desc: ""},
                    {name: "grpc_pass", desc: ""},
                    {name: "grpc_pass_header", desc: ""},
                    {name: "grpc_read_timeout", desc: ""},
                    {name: "grpc_send_timeout", desc: ""},
                    {name: "grpc_set_header", desc: ""},
                    {name: "grpc_socket_keepalive", desc: ""},
                    {name: "grpc_ssl_certificate", desc: ""},
                    {name: "grpc_ssl_certificate_key", desc: ""},
                    {name: "grpc_ssl_ciphers", desc: ""},
                    {name: "grpc_ssl_conf_command", desc: ""},
                    {name: "grpc_ssl_crl", desc: ""},
                    {name: "grpc_ssl_name", desc: ""},
                    {name: "grpc_ssl_password_file", desc: ""},
                    {name: "grpc_ssl_protocols", desc: ""},
                    {name: "grpc_ssl_server_name", desc: ""},
                    {name: "grpc_ssl_session_reuse", desc: ""},
                    {name: "grpc_ssl_trusted_certificate", desc: ""},
                    {name: "grpc_ssl_verify", desc: ""},
                    {name: "grpc_ssl_verify_depth", desc: ""},

                    {name: "image_filter", desc: ""},
                    {name: "image_filter_buffer", desc: ""},
                    {name: "image_filter_interlace", desc: ""},
                    {name: "image_filter_jpeg_quality", desc: ""},
                    {name: "image_filter_sharpen", desc: ""},
                    {name: "image_filter_transparency", desc: ""},
                    {name: "image_filter_webp_quality", desc: ""},

                    {name: "proxy_http_version", desc: "", args: ["1.1"]},
                    {name: "return", desc: "直接返回", args: ["200", "'ok'"]},
                ],
                args: [],
            },
        },
    }),
}
</script>
