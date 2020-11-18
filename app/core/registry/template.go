package registry

const default_template = `
upstream {{ .UpstreamName }} { {{range .Servers}}
	server {{.Host}}:{{.Port}} {{if ne .Weight 0}} weight={{.Weight}}{{end}};{{end}}
}

server { 
	server_name {{.Domain}};
	listen 80;

{{if .AutoSSL}}
	listen 443 ssl;
	ssl_certificate     {{.Cert.Certificate}};        
	ssl_certificate_key {{.Cert.PrivateKey}};
	ssl_session_timeout 5m;
	ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE:ECDH:AES:HIGH:!NULL:!aNULL:!MD5:!ADH:!RC4;
	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_prefer_server_ciphers on;

	if ( $scheme = 'http' ) {
       return 301 https://$host$request_uri;
    }
{{end}}

    location / {
        proxy_set_header        X-Scheme        $scheme;
        proxy_set_header        Host            $host;
        proxy_set_header        X-Real-IP       $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_pass http://{{ .UpstreamName }};
    }
}
`
