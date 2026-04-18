package entities

import (
	_ "embed"
	"unsafe"
)

//go:embed xml_decode_tree_le.bin
var xmlDecodeTreeBytes []byte

//go:embed html_decode_tree_le.bin
var htmlDecodeTreeBytes []byte

var XmlDecodeTree = embedToDecodeTree(xmlDecodeTreeBytes)
var HtmlDecodeTree = embedToDecodeTree(htmlDecodeTreeBytes)

func embedToDecodeTree(data []byte) []uint16 {
	if len(data)%2 != 0 {
		panic("The file length must be even!")
	}

	uint16Data := unsafe.Slice((*uint16)(unsafe.Pointer(&data[0])), len(data)/2)

	return uint16Data
}
