package datatypes

import (
	"io"
	"fmt"
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
		if err != nil {
			// an unknown error occurred while reading the next byte
			return value, err, totalBytesRead
		} else if bytesRead == 0 {
			// the VarInt has not ended yet but there is no more byte available
			err = io.ErrUnexpectedEOF
			return value, err, totalBytesRead
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
			return value, err, totalBytesRead
		}
		if (read & checkMask) == 0 {
			break
		}
	}
	return value, err, totalBytesRead
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
			return err, totalBytesWritten
		}
		if value == 0 {
			break
		}
	}
	return err, totalBytesWritten
}

// this method reads a VarLong from the given io.Reader
// returns the read VarLong and the amount of bytes read or an error if something went wrong
func ReadVarLong(reader io.Reader) (value int64, err error, totalBytesRead int) {
	var read byte
	singleByte := make([]byte, 1)
	for {
		bytesRead, err := reader.Read(singleByte)
		if err != nil {
			// an unknown error occurred while reading the next byte
			return value, err, totalBytesRead
		} else if bytesRead == 0 {
			// the VarInt has not ended yet but there is no more byte available
			err = io.ErrUnexpectedEOF
			return value, err, totalBytesRead
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
			return value, err, totalBytesRead
		}
		if (read & checkMask) == 0 {
			break
		}
	}
	return value, err, totalBytesRead
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
			return err, totalBytesWritten
		}
		if value == 0 {
			break
		}
	}
	return err, totalBytesWritten
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
	if err != nil {
		return value, err, totalBytesRead
	}
	totalBytesRead += prependedLengthByteAmount + length
	if length > maximumBytes || length < 1 {
		err = ErrInvalidStringLength{length}
		return value, err, totalBytesRead
	}
	buffer := make([]byte, length)
	bytesRead, err := reader.Read(buffer)
	if err != nil {
		// an unknown error occurred while reading the next byte
		return value, err, totalBytesRead
	} else if bytesRead == 0 {
		// the VarInt has not ended yet but there is no more byte available
		err = io.ErrUnexpectedEOF
		return value, err, totalBytesRead
	}
	value = string(buffer)
	return value, err, totalBytesRead
}

// this method writes a String to the given io.Writer
// returns the amount of bytes written or an error if something went wrong
func WriteString(writer io.Writer, value string) (err error, totalBytesWritten int) {
	length := len(value)
	if length > maximumBytes || length < 1 {
		err = ErrInvalidStringLength{length}
		return err, totalBytesWritten
	}
	if err, prependedLengthByteAmount := WriteVarInt(writer, length); err != nil {
		return err, totalBytesWritten
	} else {
		totalBytesWritten += prependedLengthByteAmount
	}
	stringBytes := []byte(value)
	totalBytesWritten += len(stringBytes)
	if _, err := writer.Write(stringBytes); err != nil {
		return err, totalBytesWritten
	}
	return err, totalBytesWritten
}
