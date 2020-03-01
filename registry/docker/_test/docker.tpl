## service

{{range .Services}}
    {{ $name := serviceName . }}

    {{if eq $name "web"}}
        upstream {{ $name }} {
        {{range $.Nodes}}
            server {{.Status.Addr}}:8080;
        {{end}}
        }

    {{else if eq $name "portainer_portainer"}}
        upstream {{ $name }} {
        {{ $publishPort := servicePublishedPort . 9000 }}
        {{ if eq $publishPort 0 }}
            {{ range serviceVirtualAddress . 9000 }}
                server {{.}};
            {{end}}
        {{else}}
            server {{$.PublishIP}}:{{$publishPort}};
        {{end}}
        }

    {{else if eq $name "consul_consul-server"}}
        upstream {{ $name }} {
        {{ range serviceInternalAddress . 8500 }}
            server {{.}};
        {{end}}
        }

    {{else if serviceHasLabel . "auto.domain" }}
        upstream {{ $name }} {

        }
    {{end}}

{{end}}