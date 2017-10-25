package datatypes

import (
	"bytes"
	"io"
)

// Basic class to represent sent/received packets with their id and content
type Packet struct {
	Id      int
	Content *bytes.Buffer
}

// this method writes a Packet to the given io.Writer
// returns an error if something went wrong
func WritePacket(writer io.Writer, packet Packet) (err error, totalBytesWritten int) {
	totalBuffer := bytes.NewBuffer(make([]byte, 0))
	if err, idBytesWritten := WriteVarInt(totalBuffer, packet.Id); err != nil {
		return err, totalBytesWritten
	} else {
		totalBytesWritten += idBytesWritten
	}
	if bytesWritten, err := packet.Content.WriteTo(totalBuffer); err != nil {
		return err, totalBytesWritten
	} else if bytesWritten == 0 {
		return io.ErrUnexpectedEOF, totalBytesWritten
	}
	prependedLength := len(totalBuffer.Bytes())
	if err, prependedLengthBytesWritten := WriteVarInt(writer, prependedLength); err != nil {
		return err, totalBytesWritten
	} else {
		totalBytesWritten += prependedLengthBytesWritten
	}
	if bytesWritten, err := writer.Write(totalBuffer.Bytes()); err != nil {
		return err, totalBytesWritten
	} else if bytesWritten == 0 {
		return io.ErrUnexpectedEOF, totalBytesWritten
	}
	return nil, totalBytesWritten
}

// this method reads a Packet from the given io.Reader
// returns the read Packet or an error if something went wrong
func ReadPacket(reader io.Reader) (packet Packet, err error, totalBytesRead int) {
	length, err, prependedLengthBytesRead := ReadVarInt(reader)
	if err != nil {
		return packet, err, totalBytesRead
	}
	totalBytesRead += prependedLengthBytesRead
	packetId, err, idBytesRead := ReadVarInt(reader)
	packet.Id = packetId
	if err != nil {
		return packet, err, totalBytesRead
	}
	totalBytesRead += length
	byteArray := make([]byte, length-idBytesRead)
	bytesRead, err := reader.Read(byteArray)
	if err != nil {
		return packet, err, totalBytesRead
	} else if bytesRead == 0 {
		err = io.ErrUnexpectedEOF
		return packet, err, totalBytesRead
	}
	packet.Content = bytes.NewBuffer(byteArray)
	return packet, err, totalBytesRead
}
