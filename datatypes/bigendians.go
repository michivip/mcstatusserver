package datatypes

import (
	"io"
	"encoding/binary"
)

// this method reads an unsigned short (unsigned 16-bit integer) from the given io.Reader
// returns the read unsigned short or an error if something went wrong
func ReadUnsignedShort(reader io.Reader) (value uint16, err error) {
	err = binary.Read(reader, binary.BigEndian, &value)
	return
}

// this method writes an unsigned short (unsigned 16-bit integer) to the given io.Writer
// returns an error if something went wrong
func WriteUnsignedShort(writer io.Writer, value uint16) (err error) {
	return binary.Write(writer, binary.BigEndian, &value)
}

// this method reads a long (signed 64-bit integer) from the given io.Reader
// returns the read long or an error if something went wrong
func ReadLong(reader io.Reader) (value int64, err error) {
	err = binary.Read(reader, binary.BigEndian, &value)
	return
}

// this method writes a long (signed 64-bit integer) to the given io.Writer
// returns an error if something went wrong
func WriteLong(writer io.Writer, value int64) (error) {
	return binary.Write(writer, binary.BigEndian, &value)
}
