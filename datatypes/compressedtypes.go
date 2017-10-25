package datatypes

import (
	"io"
	"fmt"
	"encoding/binary"
)

// this file contains utility methods for reading and writing the compressed data types as described here: http://wiki.vg/Protocol#Data_types

// the error which is thrown if a data type`s reading process could not be completed because of missing bytes
var InvalidTypeSize error = fmt.Errorf("the given byte sequence introduces a longer size than allowed")

// some constants for reading/writing var types
const (
	checkMask byte = 128
	mask      byte = 127
)

// this method reads a VarInt from the given io.Reader
// returns the read VarInt and the amount of bytes read or an error if something went wrong
func ReadVarInt(reader io.Reader) (value int, err error, totalBytesRead int) {
	var read byte
	singleByte := make([]byte, 1)
	for {
		bytesRead, err := reader.Read(singleByte)
		if bytesRead == 0 {
			// the VarInt has not ended yet but there is no more byte available
			err = io.ErrUnexpectedEOF
			return
		} else if err != nil {
			// an unknown error occurred while reading the next byte
			return
		}
		read = singleByte[0]
		var tempValue byte = read & mask
		var intValue int32 = int32(tempValue)
		var shiftedValue int = int(intValue << uint(7*totalBytesRead))
		value |= shiftedValue
		totalBytesRead++
		if totalBytesRead > 5 {
			// the type is bigger than allowed per definition
			err = InvalidTypeSize
			return
		}
		if (read & checkMask) == 0 {
			break
		}
	}
	return
}

// this method writes a VarInt to the given io.Writer
// returns the amount of bytes written or an error if something went wrong
func WriteVarInt(writer io.Writer, value int) (err error, totalBytesWritten int) {
	for {
		totalBytesWritten++
		var temp byte = byte(value) & mask
		value = int(uint32(value) >> 7)
		if value != 0 {
			temp |= checkMask
		}
		_, err := writer.Write([]byte{temp})
		if err != nil {
			return
		}
		if value == 0 {
			break
		}
	}
	return
}

// this method reads a VarLong from the given io.Reader
// returns the read VarLong and the amount of bytes read or an error if something went wrong
func ReadVarLong(reader io.Reader) (value int64, err error, totalBytesRead int) {
	var read byte
	singleByte := make([]byte, 1)
	for {
		bytesRead, err := reader.Read(singleByte)
		if bytesRead == 0 {
			// the VarInt has not ended yet but there is no more byte available
			err = io.ErrUnexpectedEOF
			return
		} else if err != nil {
			// an unknown error occurred while reading the next byte
			return
		}
		read = singleByte[0]
		var tempValue byte = read & mask
		var intValue int64 = int64(tempValue)
		var shiftedValue int64 = int64(intValue << uint(7*totalBytesRead))
		value |= shiftedValue
		totalBytesRead++
		if totalBytesRead > 10 {
			// the type is bigger than allowed per definition
			err = InvalidTypeSize
			return
		}
		if (read & checkMask) == 0 {
			break
		}
	}
	return
}

// this method writes a VarLong to the given io.Writer
// returns the amount of bytes written or an error if something went wrong
func WriteVarLong(writer io.Writer, value int64) (err error, totalBytesWritten int) {
	for {
		totalBytesWritten++
		var temp byte = byte(value) & mask
		value = int64(uint64(value) >> 7)
		if value != 0 {
			temp |= checkMask
		}
		_, err := writer.Write([]byte{temp})
		if err != nil {
			return
		}
		if value == 0 {
			break
		}
	}
	return
}

// some constants for reading/writing strings
const (
	maximumBytes int = 32767*4 + 3
)

// the error which is thrown if the prepended length is not valid (lower than 1 or higher than maximumBytes)
type ErrInvalidStringLength struct {
	SentLength int
}

func (stringLengthExceedError ErrInvalidStringLength) Error() string {
	return fmt.Sprintf("the prepended length (%v) is not valid (maximum: %v)", stringLengthExceedError.SentLength, maximumBytes)
}

// this method reads a String from the given io.Reader
// returns the read String and the amount of bytes read or an error if something went wrong
func ReadString(reader io.Reader) (value string, err error, totalBytesRead int) {
	length, err, prependedLengthByteAmount := ReadVarInt(reader)
	totalBytesRead += prependedLengthByteAmount + length
	if err != nil {
		return
	}
	if length > maximumBytes || length < 1 {
		err = ErrInvalidStringLength{length}
		return
	}
	buffer := make([]byte, length)
	bytesRead, err := reader.Read(buffer)
	if bytesRead == 0 {
		// the VarInt has not ended yet but there is no more byte available
		err = io.ErrUnexpectedEOF
		return
	} else if err != nil {
		// an unknown error occurred while reading the next byte
		return
	}
	value = string(buffer)
	return
}

// this method writes a String to the given io.Writer
// returns the amount of bytes written or an error if something went wrong
func WriteString(writer io.Writer, value string) (err error, totalBytesWritten int) {
	length := len(value)
	if length > maximumBytes || length < 1 {
		err = ErrInvalidStringLength{length}
		return
	}
	if err, prependedLengthByteAmount := WriteVarInt(writer, length); err != nil {
		return
	} else {
		totalBytesWritten += prependedLengthByteAmount
	}
	stringBytes := []byte(value)
	totalBytesWritten += len(stringBytes)
	if _, err := writer.Write(stringBytes); err != nil {
		return
	}
	return
}

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
