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
	mask byte = 127
)

// this method reads a VarInt from the given io.Reader
// returns the read VarInt or an error if something went wrong.
func ReadVarInt(reader io.Reader) (int, error) {
	var numRead int = 0
	var result int = 0
	var read byte
	singleByte := make([]byte, 1)
	for {
		bytesRead, err := reader.Read(singleByte)
		if bytesRead == 0 {
			// the VarInt has not ended yet but there is no more byte available
			return -1, io.ErrUnexpectedEOF
		} else if err != nil {
			// an unknown error occurred while reading the next byte
			return -1, err
		}
		read = singleByte[0]
		var value byte = read & mask
		var intValue int32 = int32(value)
		var shiftedValue int = int(intValue << uint(7*numRead))
		result |= shiftedValue
		numRead++
		if numRead > 5 {
			// the type is bigger than allowed per definition
			return -1, InvalidTypeSize
		}
		if (read & checkMask) == 0 {
			break
		}
	}
	return result, nil
}

// this method writes a VarInt to the given io.Writer
// returns an error if something went wrong.
func WriteVarInt(writer io.Writer, value int) (err error) {
	for {
		var temp byte = byte(value) & mask
		value = int(uint32(value) >> 7)
		if value != 0 {
			temp |= checkMask
		}
		_, err := writer.Write([]byte{temp})
		if err != nil {
			return err
		}
		if value == 0 {
			break
		}
	}
	return
}

// this method reads a VarLong from the given io.Reader
// returns the read VarLong or an error if something went wrong.
func ReadVarLong(reader io.Reader) (int64, error) {
	var numRead int = 0
	var result int64 = 0
	var read byte
	singleByte := make([]byte, 1)
	for {
		bytesRead, err := reader.Read(singleByte)
		if bytesRead == 0 {
			// the VarInt has not ended yet but there is no more byte available
			return -1, io.ErrUnexpectedEOF
		} else if err != nil {
			// an unknown error occurred while reading the next byte
			return -1, err
		}
		read = singleByte[0]
		var value byte = read & mask
		var intValue int64 = int64(value)
		var shiftedValue int64 = int64(intValue << uint(7*numRead))
		result |= shiftedValue
		numRead++
		if numRead > 10 {
			// the type is bigger than allowed per definition
			return -1, InvalidTypeSize
		}
		if (read & checkMask) == 0 {
			break
		}
	}
	return result, nil
}

// this method writes a VarLong to the given io.Writer
// returns an error if something went wrong.
func WriteVarLong(writer io.Writer, value int64) (err error) {
	for {
		var temp byte = byte(value) & mask
		value = int64(uint64(value) >> 7)
		if value != 0 {
			temp |= checkMask
		}
		_, err := writer.Write([]byte{temp})
		if err != nil {
			return err
		}
		if value == 0 {
			break
		}
	}
	return nil
}