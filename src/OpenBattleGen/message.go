package OpenBattleGen

// AddField represents an added field.
type AddField struct {
	Name    string `xml:"name,attr"`
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr"`
}

// RemoveField represents a removed field.
type RemoveField struct {
	Name    string `xml:"name,attr"`
	Default string `xml:"default,attr"`
}

// ChangeField represents a changed field.
type ChangeField struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

// Version represents a single version of a message.
type Version struct {
	Major   int           `xml:"major,attr"`
	Minor   int           `xml:"minor,attr"`
	Adds    []AddField    `xml:"add"`
	Removes []RemoveField `xml:"remove"`
	Changes []ChangeField `xml:"change"`
}

// Object is a class which defines a single object.
type Object struct {
	Type   string     `xml:"type,attr"`
	Fields []AddField `xml:"add"`
}

// Message is a class which defines a single message with multiple versions
// and fields.
type Message struct {
	ID       int       `xml:"id,attr"`
	Type     string    `xml:"type,attr"`
	Flags    int       `xml:"flags,attr"`
	Versions []Version `xml:"version"`
}

// Field represents a single field in a message or object.
type Field struct {
	Index int
	Name  string
	Type  string
}

// FinalVersion returns the fields of the latest version of this message.
func (msg *Message) FinalVersion() (int, map[string]Field) {
	list := make(map[string]Field)
	ver := 0
	index := 0
	for _, v := range msg.Versions {
		if v.Major > ver {
			ver = v.Major
		}
		for _, a := range v.Adds {
			list[a.Name] = Field{Index: index, Name: a.Name, Type: a.Type}
			index++
		}
		for _, r := range v.Removes {
			delete(list, r.Name)
		}
		for _, c := range v.Changes {
			list[c.Name] = Field{Index: list[c.Name].Index, Name: c.Name, Type: c.Type}
		}
	}
	return ver, list
}
