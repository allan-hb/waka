package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/liuhan907/waka/protoc/protoc-gen-cellnet/named"

	"github.com/liuhan907/waka/protoc/plugin"
	"github.com/liuhan907/waka/protoc/protoc-gen-cellnet/hash"
)

type Descriptor struct {
	ID              uint32
	FullName        string
	LeadingComments string
}

type FileDescriptor struct {
	FileName              string
	Namespace             string
	MetaProviderClassName string

	MessageType []*Descriptor
}

type Generator struct {
	*plugin.BaseGenerator
}

func (g *Generator) GenerateAllFiles() {
	for _, v := range g.Files {
		name, content := g.printFile(v)
		g.Response.File = append(g.Response.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String(name),
			Content: proto.String(content),
		})
	}
}

func (g *Generator) printFile(f *plugin.FileDescriptor) (string, string) {
	tpl, err := template.New("protoc-gen-cellnet").Parse(codeTemplate)
	if err != nil {
		g.Error(err, "template parse failed")
	}

	model := &FileDescriptor{
		FileName:              f.Descriptor.GetName(),
		Namespace:             named.BuildNamespace(f),
		MetaProviderClassName: named.BuildMetaProviderClassName(f),
	}

	for _, message := range f.MessageType {
		lastName := message.Name[len(message.Name)-1]
		if strings.HasSuffix(lastName, "Entry") {
			continue
		}
		model.MessageType = append(model.MessageType, &Descriptor{
			ID:              hash.StringHash(named.BuildCellnetFullName(f, message)),
			FullName:        named.BuildFullName(f, message),
			LeadingComments: strings.Trim(message.Location.GetLeadingComments(), "\r\n\t "),
		})
	}

	w := bytes.NewBuffer(make([]byte, 0, 1024))
	err = tpl.Execute(w, &model)
	if err != nil {
		g.Error(err, "execute template")
	}

	return fmt.Sprintf("%s.cs", named.BuildMetaProviderClassFileName(f)), w.String()
}

func NewGenerator(name string) *Generator {
	g := new(Generator)
	g.BaseGenerator = plugin.NewBaseGenerator(name)
	return g
}
