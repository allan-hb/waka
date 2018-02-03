package generator

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/liuhan907/waka/protoc/plugin"
	"github.com/liuhan907/waka/protoc/protoc-gen-cellnet/named"
)

type RPCDescriptor struct {
	Name            string
	InputType       string
	OutputType      string
	LeadingComments string
}

type TransportDescriptor struct {
	Type            string
	LeadingComments string
}

type FileDescriptor struct {
	FileName              string
	MetaProviderClassName string
	Namespace             string

	RPC     []*RPCDescriptor
	Post    []*TransportDescriptor
	Receive []*TransportDescriptor
}

type Generator struct {
	*plugin.BaseGenerator
}

func (g *Generator) GenerateAllFiles() {
	for _, v := range g.Files {
		core, supervisor, dispatcher := g.printAllFiles(v)
		g.Response.File = append(g.Response.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String("CoreSupervisor.cs"),
			Content: proto.String(core),
		})
		g.Response.File = append(g.Response.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String("Supervisor.cs"),
			Content: proto.String(supervisor),
		})
		g.Response.File = append(g.Response.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String("IDispatcher.cs"),
			Content: proto.String(dispatcher),
		})
	}
}

func (g *Generator) printAllFiles(f *plugin.FileDescriptor) (core, supervisor string, dispatcher string) {
	packages := strings.Split(f.Descriptor.GetPackage(), "_")
	for i := range packages {
		packages[i] = strings.Title(packages[i])
	}
	packageName := strings.Join(packages, "")
	model := &FileDescriptor{
		FileName:              f.Descriptor.GetName(),
		MetaProviderClassName: named.BuildMetaProviderClassName(f),
		Namespace:             packageName,
		RPC:                   g.analyseRPCs(f),
		Post:                  g.analyseTransports(f, "@post"),
		Receive:               g.analyseTransports(f, "@receive"),
	}

	return g.printFile(model, CSharpSupervisorTemplate),
		g.printFile(model, CSharpHandlerTemplate),
		g.printFile(model, CSharpIDispatcherTemplate)
}

func (g *Generator) printFile(model *FileDescriptor, tpl string) string {
	t, err := template.New("protoc-gen-waka").Parse(tpl)
	if err != nil {
		g.Error(err, "template parse failed")
	}

	w := bytes.NewBuffer(make([]byte, 0, 1024))
	err = t.Execute(w, model)
	if err != nil {
		g.Error(err, "execute template")
	}

	return w.String()
}

func (g *Generator) analyseRPCs(f *plugin.FileDescriptor) []*RPCDescriptor {
	var descriptors []*RPCDescriptor
	for _, message := range f.MessageType {
		descriptors = g.analyseRPC(descriptors, message)
	}
	return descriptors
}

func (g *Generator) analyseRPC(descriptors []*RPCDescriptor, message *plugin.Descriptor) []*RPCDescriptor {
	if message.Parent != nil {
		return descriptors
	}

	trimmed := strings.Trim(message.Location.GetLeadingComments(), "\r\n\t ")
	if trimmed == "" {
		return descriptors
	}

	var inputType string
	var outputType string
	var comments string

	split := strings.Split(trimmed, "\n")
	for _, line := range split {
		trimmed := strings.Trim(line, "\r\n\t ")
		if !strings.HasPrefix(trimmed, "@rpc") {
			comments += "        /// " + strings.Trim(strings.TrimPrefix(trimmed, "@comments"), "\r\n\t ") + "\n"
		} else {
			param := parameters(strings.Trim(strings.TrimPrefix(trimmed, "@rpc"), "\r\n\t "))
			response := param["response"]
			if response == "" {
				return descriptors
			}
			inputType = "" + message.Name[len(message.Name)-1]
			outputType = "" + response
		}
	}

	if comments == "" {
		comments = "        /// 没有注释"
	} else {
		comments = strings.Trim(comments, "\n")
	}

	if inputType == "" || outputType == "" {
		return descriptors
	}

	return append(descriptors, &RPCDescriptor{
		Name:            strings.TrimSuffix(inputType, "Request"),
		InputType:       inputType,
		OutputType:      outputType,
		LeadingComments: comments,
	})
}

func (g *Generator) analyseTransports(f *plugin.FileDescriptor, prefix string) []*TransportDescriptor {
	var descriptors []*TransportDescriptor
	for _, message := range f.MessageType {
		descriptors = g.analyseTransport(descriptors, message, prefix)
	}
	return descriptors
}

func (g *Generator) analyseTransport(descriptors []*TransportDescriptor, message *plugin.Descriptor, prefix string) []*TransportDescriptor {
	if message.Parent != nil {
		return descriptors
	}

	trimmed := strings.Trim(message.Location.GetLeadingComments(), "\r\n\t ")
	if trimmed == "" {
		return descriptors
	}

	var typeName string
	var comments string

	split := strings.Split(trimmed, "\n")
	for _, line := range split {
		trimmed := strings.Trim(line, "\r\n\t ")
		if strings.HasPrefix(trimmed, prefix) {
			typeName = "" + message.Name[len(message.Name)-1]
		} else {
			comments += "        /// " + strings.Trim(strings.TrimPrefix(trimmed, "@comments"), "\r\n\t ") + "\n"
		}
	}

	if comments == "" {
		comments = "        /// 没有注释"
	} else {
		comments = strings.Trim(comments, "\n")
	}

	if typeName == "" {
		return descriptors
	}

	return append(descriptors, &TransportDescriptor{
		Type:            typeName,
		LeadingComments: comments,
	})
}

func NewGenerator(name string) *Generator {
	g := new(Generator)
	g.BaseGenerator = plugin.NewBaseGenerator(name)
	return g
}
