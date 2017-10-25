package datatypes

import (
	"bytes"
	"io"
)

type PacketId int

func (packetId PacketId) Abs() int {
	return int(packetId)
}

// this method writes a Packet to the given io.Writer
// returns an error if something went wrong
func WritePacket(writer io.Writer, packetId PacketId, data bytes.Buffer) (err error) {
	totalBuffer := bytes.NewBuffer(make([]byte, 0))
	if err := WriteVarInt(totalBuffer, packetId.Abs()); err != nil {
		return
	}
	if bytesWritten, err := data.WriteTo(totalBuffer); err != nil {
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
