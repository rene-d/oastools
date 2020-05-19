package {{.PackageName}}
// GENERATED DO NOT EDIT
{{ $package := .PackageName }}
import(
	"log"
    "github.com/Axili39/gowapi"
    "google.golang.org/protobuf/proto"
)

{{- range $val := .WSHandlers }}
type {{$val.Name}}ClientOps interface {
    {{- range $op := $val.ClientInterface.Ops}}
	op{{$op.Name}}(*{{$op.PbMessageName}})
    {{- end }}
}

type {{$val.Name}}ServerPeer struct {
	conn *gowapi.Conn
    ops {{$val.Name}}ClientOps
}

func (p *{{$val.Name}}ServerPeer) OnRecvMessage(message []byte) error {
	var operation {{$val.ClientInterface.ContainerMessage}}
	err := proto.Unmarshal(message, &operation)

	switch operation.Select.(type) {
    {{- range $op := $val.ClientInterface.Ops}}    
	case *{{$val.ClientInterface.ContainerMessage}}_{{$op.PbMessageName}}Value:
		p.ops.op{{$op.Name}}(operation.Get{{$op.PbMessageName}}Value())
    {{- end }}
	default:
		log.Printf(" unrecognized %v, discarderd\n", operation)
	}
	return err
}

// Helpers for calling Server Interface
{{- range $op := $val.ServerInterface.Ops}}    
func (p *{{$val.Name}}ServerPeer) Invoke{{$op.Name}}(msg *{{$op.PbMessageName}}) error {
    var container {{$val.ServerInterface.ContainerMessage}}
	container.Select = &{{$val.ServerInterface.ContainerMessage}}_{{$op.PbMessageName}}Value{msg}
	
    return p.conn.ProtoWrite(&container)
}
{{- end }}
func (p *{{$val.Name}}ServerPeer) OnClose() {
	log.Printf(" closing....   %v", p)
}

func DialServer(addr string, ops {{$val.Name}}ClientOps)  (*{{$val.Name}}ServerPeer, error) {
	var err error
	peer := {{$val.Name}}ServerPeer{nil, ops}

	peer.conn, _, err = gowapi.Dial("ws://" + addr + "{{$val.Path}}", nil, &peer)
   
    if err != nil {
		log.Printf("error establishing websocket %v\n", err)
		return nil, err
	}
	return &peer,nil
}
{{ end }}