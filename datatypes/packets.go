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
	totalBuffer := bytes.NewBuffer([]byte{})
	dataBuffer := bytes.NewBuffer([]byte{})
	if err, _ := WriteVarInt(dataBuffer, packet.Id); err != nil {
		return err, totalBytesWritten
	}
	if bytesWritten, err := dataBuffer.Write(packet.Content.Bytes()); err != nil {
		return err, totalBytesWritten
	} else if bytesWritten < packet.Content.Len() {
		return io.ErrUnexpectedEOF, totalBytesWritten
	}
	if err, bytesWritten := WriteVarInt(totalBuffer, dataBuffer.Len()); err != nil {
		return err, totalBytesWritten
	} else {
		totalBytesWritten += bytesWritten
	}
	totalBuffer.Write(dataBuffer.Bytes())
	if bytesWritten, err := writer.Write(totalBuffer.Bytes()); err != nil {
		return err, totalBytesWritten
	} else if int(bytesWritten) < totalBuffer.Len() {
		return io.ErrUnexpectedEOF, totalBytesWritten
	} else {
		totalBytesWritten += int(bytesWritten)
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
	if len(byteArray) > 0 {
		bytesRead, err := reader.Read(byteArray)
		if err != nil {
			return packet, err, totalBytesRead
		} else if bytesRead < len(byteArray) {
			err = io.ErrUnexpectedEOF
			return packet, err, totalBytesRead
		}
	}
	packet.Content = bytes.NewBuffer(byteArray)
	return packet, err, totalBytesRead
}
