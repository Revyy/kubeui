/* Code copied and changed from the colorjson library to work with lipgloss.
* https://github.com/TylerBrock/colorjson
* Attribution to Tyler Brock!
 */

package jsoncolor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const initialDepth = 0
const valueSep = ","
const null = "null"
const startMap = "{"
const endMap = "}"
const startArray = "["
const endArray = "]"

const emptyMap = startMap + endMap
const emptyArray = startArray + endArray

// Formatter provides configuration and marshaling of json objects into colored json strings.
type Formatter struct {
	ColorSet
	StringMaxLength int
	Indent          int
	RawStrings      bool
}

// ColorSet defined the color of the json data types.
type ColorSet struct {
	KeyColor    lipgloss.Color
	StringColor lipgloss.Color
	BoolColor   lipgloss.Color
	NumberColor lipgloss.Color
	NullColor   lipgloss.Color
}

func defaultColorSet() ColorSet {
	return ColorSet{
		KeyColor:    lipgloss.Color("#ffffff"),
		StringColor: lipgloss.Color("#438a34"),
		BoolColor:   lipgloss.Color("#c4cc23"),
		NumberColor: lipgloss.Color("#19d4ae"),
		NullColor:   lipgloss.Color("#ba11a6"),
	}
}

func lightColorSet() ColorSet {
	return ColorSet{
		KeyColor:    lipgloss.Color("#333333"),
		StringColor: lipgloss.Color("#438a34"),
		BoolColor:   lipgloss.Color("#c4cc23"),
		NumberColor: lipgloss.Color("#19d4ae"),
		NullColor:   lipgloss.Color("#ba11a6"),
	}
}

// NewFormatter creates a new formatter.
func NewFormatter() *Formatter {

	colorSet := defaultColorSet()

	if !lipgloss.HasDarkBackground() {
		colorSet = lightColorSet()
	}

	return &Formatter{
		ColorSet:        colorSet,
		StringMaxLength: 0,
		Indent:          0,
		RawStrings:      false,
	}
}

func (f *Formatter) sprintColor(c lipgloss.Color, s string) string {
	return lipgloss.NewStyle().Foreground(c).Render(s)
}

func (f *Formatter) writeIndent(buf *bytes.Buffer, depth int) {
	buf.WriteString(strings.Repeat(" ", f.Indent*depth))
}

func (f *Formatter) writeObjSep(buf *bytes.Buffer) {
	if f.Indent != 0 {
		buf.WriteByte('\n')
	} else {
		buf.WriteByte(' ')
	}
}

// Marshal marshals the object into a colored json string.
func (f *Formatter) Marshal(jsonObj interface{}) ([]byte, error) {
	buffer := bytes.Buffer{}
	f.marshalValue(jsonObj, &buffer, initialDepth)
	return buffer.Bytes(), nil
}

func (f *Formatter) marshalMap(m map[string]interface{}, buf *bytes.Buffer, depth int) {
	remaining := len(m)

	if remaining == 0 {
		buf.WriteString(emptyMap)
		return
	}

	keys := make([]string, 0)
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	buf.WriteString(startMap)
	f.writeObjSep(buf)

	for _, key := range keys {
		f.writeIndent(buf, depth+1)
		buf.WriteString(f.sprintColor(f.KeyColor, fmt.Sprintf("\"%s\": ", key)))
		f.marshalValue(m[key], buf, depth+1)
		remaining--
		if remaining != 0 {
			buf.WriteString(valueSep)
		}
		f.writeObjSep(buf)
	}
	f.writeIndent(buf, depth)
	buf.WriteString(endMap)
}

func (f *Formatter) marshalArray(a []interface{}, buf *bytes.Buffer, depth int) {
	if len(a) == 0 {
		buf.WriteString(emptyArray)
		return
	}

	buf.WriteString(startArray)
	f.writeObjSep(buf)

	for i, v := range a {
		f.writeIndent(buf, depth+1)
		f.marshalValue(v, buf, depth+1)
		if i < len(a)-1 {
			buf.WriteString(valueSep)
		}
		f.writeObjSep(buf)
	}
	f.writeIndent(buf, depth)
	buf.WriteString(endArray)
}

func (f *Formatter) marshalValue(val interface{}, buf *bytes.Buffer, depth int) {
	switch v := val.(type) {
	case map[string]interface{}:
		f.marshalMap(v, buf, depth)
	case []interface{}:
		f.marshalArray(v, buf, depth)
	case string:
		f.marshalString(v, buf)
	case float64:
		buf.WriteString(f.sprintColor(f.NumberColor, strconv.FormatFloat(v, 'f', -1, 64)))
	case bool:
		buf.WriteString(f.sprintColor(f.BoolColor, (strconv.FormatBool(v))))
	case nil:
		buf.WriteString(f.sprintColor(f.NullColor, null))
	case json.Number:
		buf.WriteString(f.sprintColor(f.NumberColor, v.String()))
	}
}

func (f *Formatter) marshalString(str string, buf *bytes.Buffer) {
	if !f.RawStrings {
		strBytes, _ := json.Marshal(str)
		str = string(strBytes)
	}

	if f.StringMaxLength != 0 && len(str) >= f.StringMaxLength {
		str = fmt.Sprintf("%s...", str[0:f.StringMaxLength])
	}

	buf.WriteString(f.sprintColor(f.StringColor, str))
}

// Marshal marshals JSON data with default options.
func Marshal(jsonObj interface{}) ([]byte, error) {
	return NewFormatter().Marshal(jsonObj)
}
