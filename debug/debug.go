package debug

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"layeh.com/radius"
	"layeh.com/radius/dictionary"
)

type Config struct {
	Dictionary *dictionary.Dictionary
}

func DumpPacket(c *Config, p *radius.Packet) string {
	var b bytes.Buffer
	b.WriteString(p.Code.String())
	b.WriteString(" Id ")
	b.WriteString(strconv.Itoa(int(p.Identifier)))
	b.WriteByte('\n')

	for _, elem := range sortedAttributes(p.Attributes) {
		attrsType, attrs := elem.Type, elem.Attrs
		if len(attrs) == 0 {
			continue
		}

		attrsTypeIntStr := strconv.Itoa(int(attrsType))
		var attrTypeString string
		var stringerFunc func(p *radius.Packet, attr *dictionary.Attribute, value []byte) string
		dictAttr := c.Dictionary.AttributeByOID(attrsTypeIntStr)
		if dictAttr != nil {
			attrTypeString = dictAttr.Name
			switch dictAttr.Type {
			case dictionary.AttributeString, dictionary.AttributeOctets:
				stringerFunc = maybetextStringer
			case dictionary.AttributeInteger:
				stringerFunc = intStringer
			default:
				stringerFunc = hexStringer
			}
		} else {
			attrTypeString = "#" + attrsTypeIntStr
			stringerFunc = hexStringer
		}

		for _, attr := range attrs {
			b.WriteString("  ")
			b.WriteString(attrTypeString)
			b.WriteString(" = ")
			b.WriteString(stringerFunc(p, dictAttr, attr))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func intStringer(p *radius.Packet, attr *dictionary.Attribute, value []byte) string {
	switch len(value) {
	case 4:
		return strconv.Itoa(int(binary.BigEndian.Uint32(value)))
	case 8:
		return strconv.Itoa(int(binary.BigEndian.Uint64(value)))
	}
	return "0x" + hex.EncodeToString(value)
}

func maybetextStringer(p *radius.Packet, attr *dictionary.Attribute, value []byte) string {
	if attr != nil && attr.FlagEncrypt != nil && *attr.FlagEncrypt == 1 {
		decryptedValue, err := radius.UserPassword(radius.Attribute(value), p.Secret, p.Authenticator[:])
		if err != nil {
			return "0x" + hex.EncodeToString(value)
		}
		value = decryptedValue
	}
	return fmt.Sprintf("%q", value)
}

func hexStringer(p *radius.Packet, attr *dictionary.Attribute, value []byte) string {
	return "0x" + hex.EncodeToString(value)
}

type attributesElement struct {
	Type  radius.Type
	Attrs []radius.Attribute
}

func sortedAttributes(attributes radius.Attributes) []attributesElement {
	var sortedAttrs []attributesElement
	for attrsType, attrs := range attributes {
		sortedAttrs = append(sortedAttrs, attributesElement{
			Type:  attrsType,
			Attrs: attrs,
		})
	}

	sort.Sort(sortAttributesType(sortedAttrs))

	return sortedAttrs
}

type sortAttributesType []attributesElement

func (s sortAttributesType) Len() int           { return len(s) }
func (s sortAttributesType) Less(i, j int) bool { return s[i].Type < s[j].Type }
func (s sortAttributesType) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
