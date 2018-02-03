package named

import (
	"fmt"
	"strings"

	"github.com/liuhan907/waka/protoc/plugin"
)

func BuildMetaProviderClassFileName(f *plugin.FileDescriptor) string {
	files := strings.Split(strings.TrimSuffix(f.Descriptor.GetName(), ".proto"), "_")
	for i := range files {
		files[i] = strings.Title(files[i])
	}
	return fmt.Sprintf("%sMetaProvider", strings.Join(files, ""))
}

func BuildMetaProviderClassName(f *plugin.FileDescriptor) string {
	packages := strings.Split(f.Descriptor.GetPackage(), "_")
	for i := range packages {
		packages[i] = strings.Title(packages[i])
	}
	return fmt.Sprintf("%sMetaProvider", strings.Join(packages, ""))
}

func BuildCellnetFullName(f *plugin.FileDescriptor, message *plugin.Descriptor) string {
	return fmt.Sprintf("%s.%s", f.Descriptor.GetPackage(), strings.Join(message.Name, "."))
}

func BuildFullName(f *plugin.FileDescriptor, message *plugin.Descriptor) string {
	packages := strings.Split(f.Descriptor.GetPackage(), "_")
	for i := range packages {
		packages[i] = strings.Title(packages[i])
	}
	return fmt.Sprintf("%s.%s", strings.Join(packages, ""), strings.Join(message.Name, ".Types."))
}

func BuildNamespace(f *plugin.FileDescriptor) string {
	packages := strings.Split(f.Descriptor.GetPackage(), "_")
	for i := range packages {
		packages[i] = strings.Title(packages[i])
	}
	return strings.Join(packages, "")
}
