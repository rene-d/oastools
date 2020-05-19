package {{.PackageName}}
// GENERATED DO NOT EDIT

import (
	"log"
	"net/http"
	"github.com/Axili39/gowapi"
	"google.golang.org/protobuf/proto"
)

{{ $package := .PackageName }}
type Server interface {
	{{- range $val := .HTTPHandlers }}
	{{$val.Operation}}(w http.ResponseWriter, r *http.Request)
    {{- end }}
	{{- range $val := .WSHandlers }}
	Get{{$val.Name}}() gowapi.WSHandler
    {{- end }}
}

func RegisterPaths(router *gowapi.RouteServer, s Server) {
    {{- range $val := .HTTPHandlers }}
    router.AddHandler("{{$val.Path}}", "{{$val.Method}}", s.{{$val.Operation}})
    {{- end }}
    {{- range $val := .WSHandlers }}
    router.AddWsHandler("{{$val.Path}}", s.Get{{$val.Name}}())
    {{- end }}
}

{{- range $val := .WSHandlers }}
type {{$val.Name}}ServerOps interface {
    {{- range $op := $val.ServerInterface.Ops}}
	op{{$op.Name}}(*{{$op.PbMessageName}})
    {{- end }}
}

type {{$val.Name}}ClientPeer struct {
	conn *gowapi.Conn
    ops {{$val.Name}}ServerOps
}

func {{$val.Name}}CreatePeer(c *gowapi.Conn, ops {{$val.Name}}ServerOps) *{{$val.Name}}ClientPeer {
	return &{{$val.Name}}ClientPeer{c, ops}
}

func (p *{{$val.Name}}ClientPeer) OnRecvMessage(message []byte) error {
	var operation {{$val.ServerInterface.ContainerMessage}}
	err := proto.Unmarshal(message, &operation)

	switch operation.Select.(type) {
    {{- range $op := $val.ServerInterface.Ops}}    
	case *{{$val.ServerInterface.ContainerMessage}}_{{$op.PbMessageName}}Value:
		p.ops.op{{$op.Name}}(operation.Get{{$op.PbMessageName}}Value())
    {{- end }}
	default:
		log.Printf(" unrecognized %v, discarderd\n", operation)
	}
	return err
}

{{- range $op := $val.ClientInterface.Ops}}    
func (p *{{$val.Name}}ClientPeer) Invoke{{$op.Name}}(msg *{{$op.PbMessageName}}) error {
    var container {{$val.ClientInterface.ContainerMessage}}
	container.Select = &{{$val.ClientInterface.ContainerMessage}}_{{$op.PbMessageName}}Value{msg}
	
    return p.conn.ProtoWrite(&container)
}

{{- end }}

func (p *{{$val.Name}}ClientPeer) OnClose() {
	log.Printf(" closing....   %v", p)
}

{{ end }}