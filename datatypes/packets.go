package datatypes

import (
	"bytes"
	"io"
)

// Basic class to represent sent/received packets with their id and content
type Packet struct {
	Id      int
	Content bytes.Buffer
}

// this method writes a Packet to the given io.Writer
// returns an error if something went wrong
func WritePacket(writer io.Writer, packet Packet) (err error) {
	totalBuffer := bytes.NewBuffer(make([]byte, 0))
	if err := WriteVarInt(totalBuffer, packet.Id); err != nil {
		return
	}
	if bytesWritten, err := packet.Content.WriteTo(totalBuffer); err != nil {
		return
	} else if bytesWritten == 0 {
		return io.ErrUnexpectedEOF
	}
	prependedLength := len(totalBuffer.Bytes())
	if err := WriteVarInt(writer, prependedLength); err != nil {
		return
	}
	if bytesWritten, err := writer.Write(totalBuffer.Bytes()); err != nil {
		return
	} else if bytesWritten == 0 {
		return io.ErrUnexpectedEOF
	}
	return nil
}
