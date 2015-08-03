package OpenBattleGen

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Definitions is a type which contains message definitions.
type Definitions struct {
	Objects  []Object
	Messages []Message
}

// Parse parses the input XML file for message definitions.
func (d *Definitions) Parse(input string) error {
	type openBattle struct {
		XMLName  xml.Name  `xml:"OpenBattle"`
		Includes []string  `xml:"definition"`
		Objects  []Object  `xml:"object"`
		Messages []Message `xml:"message"`
	}

	p := filepath.Dir(input)
	files := []string{input}
	for i := 0; i < len(files); i++ {
		fullpath := files[i]
		if strings.Index(fullpath, p) != 0 {
			fullpath = filepath.Join(p, fullpath)
		}
		fmt.Println("File:", fullpath)
		file, err := os.Open(fullpath)
		if err != nil {
			return err
		}

		var defs openBattle
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		xml.Unmarshal(data, &defs)

		files = append(files, defs.Includes...)
		d.Objects = append(d.Objects, defs.Objects...)
		d.Messages = append(d.Messages, defs.Messages...)
	}

	return nil
}
