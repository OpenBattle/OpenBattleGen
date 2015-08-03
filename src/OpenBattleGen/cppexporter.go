package OpenBattleGen

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func init() {
	RegisterExporter(new(CppExporter))
}

var (
	builtin = []string{"long", "unsigned long", "int", "unsigned int", "short", "unsigned short", "char", "unsigned char", "float", "String"}
)

// CppExporter exports message information in C++ format.
type CppExporter struct {
}

// Languages returns an array of supported output languages.
func (exp *CppExporter) Languages() []string {
	return []string{"c", "cpp"}
}

// Export exports the specified definitions using this exporter.
func (exp *CppExporter) Export(d *Definitions, w io.Writer) error {
	fmt.Fprintln(w, "#ifndef _GENERATED")
	fmt.Fprintln(w, "#define _GENERATED")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "#include \"OpenBattleCore.h\"")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "namespace OpenBattle {")

	for _, obj := range d.Objects {
		if err := exp.exportObject(&obj, w, "\t"); err != nil {
			return err
		}
	}

	for _, msg := range d.Messages {
		if err := exp.exportMessage(&msg, w, "\t"); err != nil {
			return err
		}
	}

	fmt.Fprintln(w, "#define REGISTER(p)\t\\")
	for _, msg := range d.Messages {
		fmt.Fprintf(w, "\tp.addFactory(new %sMessageFactory());\t\\\n", msg.Type)
	}
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "#endif")
	fmt.Fprintln(w, "")
	return nil
}

type field struct {
	name     string
	index    int
	dataType dataType
}

type dataType struct {
	orig       string
	class      string
	ref        string
	template   string
	array      bool
	builtin    bool
	disposable bool
	pointer    int
}

type byIndex []field

func (a byIndex) Len() int           { return len(a) }
func (a byIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byIndex) Less(i, j int) bool { return a[i].index < a[j].index }

func getDataType(val string) dataType {
	d := dataType{orig: val}
	for _, s := range builtin {
		if i := strings.Index(val, s); i > -1 {
			d.builtin = true
			d.class = s
			val = strings.Replace(val, s, "", 1)
			break
		}
	}
	if i := strings.Index(val, "<"); i > -1 {
		if j := strings.Index(val, ">"); j > i {
			d.template = val[i+1 : j]
			val = val[:i] + val[j+1:]
		}
	}
	if i := strings.Index(val, "[]"); i > -1 {
		d.array = true
		d.disposable = true
		val = strings.Replace(val, "[]", "", 1)
	}
	for strings.Index(val, "*") > -1 {
		d.pointer++
		d.disposable = true
		val = strings.Replace(val, "*", "", 1)
	}
	if d.template != "" {
		d.class = fmt.Sprintf("%s<%s>", val, d.template)
	}
	if d.class == "" {
		d.class = strings.TrimSpace(val)
	}
	d.ref = d.class
	if d.pointer > 0 {
		d.ref = d.class + strings.Repeat("*", d.pointer)
	}
	return d
}

func finalVersion(msg *Message) (int, []field) {
	ret := []field{}
	ver, final := msg.FinalVersion()
	for key := range final {
		ret = append(ret, field{
			name:     key,
			dataType: getDataType(final[key].Type),
		})
	}
	return ver, ret
}

func writeDeserializer(w io.Writer, final []field, indent string) {
	fmt.Fprintln(w, indent+"bool deserialize(Serializer *serializer) {")
	for _, f := range final {
		if f.dataType.builtin {
			fmt.Fprintf(w, indent+"\tif (this->%s == 0 && !serializer->read(&this->%s)) {\n", f.name, f.name)
			fmt.Fprintln(w, indent+"\t\treturn false;")
			fmt.Fprintln(w, indent+"\t}")
		} else if f.dataType.array {
			fmt.Fprintf(w, indent+"\tif (this->%s == 0) {\n", f.name)
			fmt.Fprintln(w, indent+"\t\tunsigned char __len = 0;")
			fmt.Fprintln(w, indent+"\t\tif (!serializer->read(&__len)) {")
			fmt.Fprintln(w, indent+"\t\t\treturn false;")
			fmt.Fprintln(w, indent+"\t\t}")
			fmt.Fprintf(w, indent+"\t\tthis->%s = new %s[__len];\n", f.name, f.dataType.class)
			// fmt.Fprintln(w, indent+"\t\tfor (unsigned char __i = 0; __i < __len; __i++) {")
			// fmt.Fprintf(w, indent+"\t\t\tthis->%s[__i] = 0;\n", f.name)
			// fmt.Fprintln(w, indent+"\t\t}")
			fmt.Fprintln(w, indent+"\t}")
			fmt.Fprintf(w, indent+"\tunsigned char __%slen = sizeof(this->%s) / sizeof(%s);\n", f.name, f.name, f.dataType.class)
			fmt.Fprintf(w, indent+"\tfor (unsigned char __i = 0; __i < __%slen; __i++) {\n", f.name)
			// fmt.Fprintf(w, indent+"\t\tif (this->%s[__i] == 0) {\n", f.name)
			// fmt.Fprintf(w, indent+"\t\t\tthis->%s[__i] = %s();\n", f.name, f.dataType.class)
			// fmt.Fprintln(w, indent+"\t\t}")
			fmt.Fprintf(w, indent+"\t\tif (!this->%s[__i].deserialize(serializer)) {\n", f.name)
			fmt.Fprintln(w, indent+"\t\t\treturn false;")
			fmt.Fprintln(w, indent+"\t\t}")
			fmt.Fprintln(w, indent+"\t}")
		} else if f.dataType.pointer > 0 {
			fmt.Fprintf(w, indent+"\tif (this->%s == 0) {\n", f.name)
			fmt.Fprintf(w, indent+"\t\tthis->%s = new %s();\n", f.name, f.dataType.class)
			fmt.Fprintln(w, indent+"\t}")
			fmt.Fprintf(w, indent+"\tif (!this->%s->deserialize(serializer)) {\n", f.name)
			fmt.Fprintln(w, indent+"\t\treturn false;")
			fmt.Fprintln(w, indent+"\t}")
		} else {
			fmt.Fprintf(w, indent+"\tif (!this->%s.deserialize(serializer)) {\n", f.name)
			fmt.Fprintln(w, indent+"\t\treturn false;")
			fmt.Fprintln(w, indent+"\t}")
		}
	}
	fmt.Fprintln(w, indent+"\treturn true;")
	fmt.Fprintln(w, indent+"}")
}

func writeSerializer(w io.Writer, final []field, indent string, msg bool) {
	fmt.Fprintln(w, indent+"void serialize(Serializer *serializer, Stream *stream) {")
	if msg {
		fmt.Fprintln(w, indent+"\tMessage::serialize(serializer, stream);")
	}
	for _, f := range final {
		if f.dataType.builtin {
			fmt.Fprintf(w, indent+"\tserializer->write(this->%s, stream);\n", f.name)
		} else if f.dataType.array {
			fmt.Fprintf(w, indent+"\tunsigned char __%slen = sizeof(this->%s) / sizeof(%s);\n", f.name, f.name, f.dataType.class)
			fmt.Fprintf(w, indent+"\tserializer->write(__%slen, stream);\n", f.name)
			fmt.Fprintf(w, indent+"\tfor (unsigned char __i = 0; __i < __%slen; __i++) {\n", f.name)
			fmt.Fprintf(w, indent+"\t\tthis->%s[__i].serialize(serializer, stream);\n", f.name)
			fmt.Fprintln(w, indent+"\t}")
		} else if f.dataType.pointer > 0 {
			fmt.Fprintf(w, indent+"\tthis->%s->serialize(serializer, stream);\n", f.name)
		} else {
			fmt.Fprintf(w, indent+"\tthis->%s.serialize(serializer, stream);\n", f.name)
		}
	}
	fmt.Fprintln(w, indent+"}")
}

func writeLength(w io.Writer, final []field, indent string) {
	fmt.Fprintln(w, indent+"unsigned char length(Serializer* serializer) {")
	fmt.Fprintln(w, indent+"\tunsigned char len = 0;")
	for _, f := range final {
		if f.dataType.disposable {
			fmt.Fprintf(w, indent+"\tlen += serializer->sizeOf(this->%s);\n", f.name)
		} else {
			fmt.Fprintf(w, indent+"\tlen += sizeof this->%s;\n", f.name)
		}
	}
	fmt.Fprintln(w, indent+"\treturn len;")
	fmt.Fprintln(w, indent+"}")
}

func writeDispose(w io.Writer, disposables []field, indent string) {
	fmt.Fprintln(w, indent+"void dispose() {")
	for _, field := range disposables {
		fmt.Fprintf(w, indent+"\tif (this->%s != 0) {\n", field.name)
		if field.dataType.array {
			fmt.Fprintf(w, indent+"\t\tfor (unsigned char __i = 0; __i < sizeof(this->%s) / sizeof(%s); __i++) {\n", field.name, field.dataType.ref)
			fmt.Fprintf(w, indent+"\t\t\tthis->%s[__i].dispose();\n", field.name)
			fmt.Fprintln(w, indent+"\t\t}")
		}
		fmt.Fprintf(w, indent+"\t\tdelete this->%s;\n", field.name)
		fmt.Fprintf(w, indent+"\t\tthis->%s = 0;\n", field.name)
		fmt.Fprintln(w, indent+"\t}")
	}
	fmt.Fprintln(w, indent+"}")
}

func writeDeclarations(w io.Writer, final []field, indent string) []field {
	disposables := []field{}
	for _, f := range final {
		if f.dataType.disposable {
			disposables = append(disposables, f)
		}
		dt := f.dataType.ref
		if f.dataType.builtin {
			dt = f.dataType.orig
		}
		fmt.Fprintf(w, indent+"%s %s = 0;\n", dt, f.name)
	}
	return disposables
}

func (exp *CppExporter) exportObject(obj *Object, w io.Writer, indent string) error {
	fields := []field{}
	for _, f := range obj.Fields {
		fields = append(fields, field{
			name:     f.Name,
			dataType: getDataType(f.Type),
		})
	}

	fmt.Fprintf(w, indent+"class %s : public ISerializable {\n", obj.Type)
	fmt.Fprintln(w, indent+"public:")

	if len(fields) > 0 {
		ctor := "\t" + obj.Type + "("
		params := ") : "
		defaults := ""
		for i, f := range fields {
			ctor += f.dataType.ref + " " + f.name
			params += f.name + "(" + f.name + ")"
			defaults += f.name + "(0)"
			if i < len(fields)-1 {
				ctor += ", "
				params += ", "
				defaults += ", "
			}
		}

		fmt.Fprintf(w, indent+"\t%s() : %s {}\n", obj.Type, defaults)
		fmt.Fprintf(w, indent+"%s%s {}\n", ctor, params)
	}

	disposables := writeDeclarations(w, fields, indent+"\t")
	fmt.Fprintln(w, "")

	writeLength(w, fields, indent+"\t")
	writeDeserializer(w, fields, indent+"\t")
	writeSerializer(w, fields, indent+"\t", false)
	writeDispose(w, disposables, indent+"\t")

	fmt.Fprintln(w, indent+"};")
	fmt.Fprintln(w, "")

	return nil
}

func (exp *CppExporter) exportMessage(msg *Message, w io.Writer, indent string) error {
	ver, final := finalVersion(msg)
	sort.Sort(byIndex(final))

	fmt.Fprintf(w, indent+"#define MSG_%s 0x%x\n", strings.ToUpper(msg.Type), msg.ID)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, indent+"class %sMessage : public Message {\n", msg.Type)
	fmt.Fprintln(w, indent+"\tunsigned char ver;")
	fmt.Fprintln(w, indent+"public:")
	fmt.Fprintf(w, indent+"\t%sMessage() : ver(0x%x) {}\n", msg.Type, ver)
	fmt.Fprintf(w, indent+"\t%sMessage(unsigned char ver) : ver(ver) {}\n", msg.Type)

	disposables := writeDeclarations(w, final, indent+"\t")
	fmt.Fprintln(w, "")

	fmt.Fprintf(w, indent+"\tunsigned char type() { return 0x%x; }\n", msg.ID)
	fmt.Fprintln(w, indent+"\tunsigned char version() { return this->ver; }")

	writeLength(w, final, indent+"\t")
	writeDeserializer(w, final, indent+"\t")
	writeSerializer(w, final, indent+"\t", true)
	writeDispose(w, disposables, indent+"\t")

	fmt.Fprintln(w, indent+"};")

	fmt.Fprintf(w, indent+"class %sMessageFactory : public MessageFactory {\n", msg.Type)
	fmt.Fprintln(w, indent+"public:")
	fmt.Fprintf(w, indent+"\tunsigned char type() { return 0x%x; }\n", msg.ID)
	fmt.Fprintf(w, indent+"\tMessage* get(unsigned char ver) { return new %sMessage(ver); }\n", msg.Type)
	fmt.Fprintln(w, indent+"};")

	fmt.Fprintln(w, "")

	return nil
}
