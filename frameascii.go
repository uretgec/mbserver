package mbserver

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// ASCIIFrame is the Modbus TCP frame.
type ASCIIFrame struct {
	StartComon uint8 // :
	Address    uint8
	Function   uint8
	Data       []byte
	LRC        uint16
	CR         uint8 // \r
	LF         uint8 // \n
}

// NewASCIIFrame converts a packet to a Modbus TCP frame.
func NewASCIIFrame(packet []byte) (*ASCIIFrame, error) {
	// Check the that the packet length.
	if len(packet) < 5 {
		return nil, fmt.Errorf("ASCII Frame error: packet less than 5 bytes: %v", packet)
	}

	// LRC
	pLen := len(packet)
	data := packet[1 : pLen-2]

	// packet frame decode
	decodeData := make([]byte, len(data)/2)
	length, err := hex.Decode(decodeData, data)
	if err != nil {
		err := fmt.Errorf("modbus: response data wrong")
		return nil, err
	}

	// lrc caclulate
	lrcExpect := decodeData[length-1]
	lrcCalc := lrcModbus(decodeData[:length-1]...) // remove last byte (lrc)
	if lrcExpect != lrcCalc {
		err := fmt.Errorf("modbus: response lrc '%v' does not match expected '%v'", lrcCalc, lrcExpect)
		return nil, err
	}

	frame := &ASCIIFrame{
		Address:  uint8(decodeData[0]),
		Function: uint8(decodeData[1]),
		Data:     decodeData[2 : length-1],
	}

	return frame, nil
}

// Copy the ASCIIFrame.
func (frame *ASCIIFrame) Copy() Framer {
	copy := *frame
	return &copy
}

// Bytes returns the Modbus byte stream based on the ASCIIFrame fields
func (frame *ASCIIFrame) Bytes() []byte {
	byts := make([]byte, 2)

	byts[0] = frame.Address
	byts[1] = frame.Function
	byts = append(byts, frame.Data...)

	// Calculate the LRC. 1 byte
	pLen := len(byts)
	lrc := lrcModbus(byts[:pLen]...)

	// Add the LRC.
	byts = append(byts, lrc)

	pLen += 1

	asciiTotalBytes := pLen * 2

	ascii_packet := make([]byte, asciiTotalBytes+3) // +3 = :, CR + LF
	_ = hex.Encode(ascii_packet[1:], byts[:pLen])

	asciiTotalBytes += 1

	// start ascii
	ascii_packet[0] = 0x3a // 1 byte (:)

	// end line: CR LF
	ascii_packet[asciiTotalBytes] = 0x0d   // CR (\r)
	ascii_packet[asciiTotalBytes+1] = 0x0a // LF (\n)

	asciiTotalBytes += 2

	return bytes.ToUpper(ascii_packet[:asciiTotalBytes])
}

// GetFunction returns the Modbus function code.
func (frame *ASCIIFrame) GetFunction() uint8 {
	return frame.Function
}

// GetData returns the ASCIIFrame Data byte field.
func (frame *ASCIIFrame) GetData() []byte {
	return frame.Data
}

// SetData sets the ASCIIFrame Data byte field and updates the frame length
// accordingly.
func (frame *ASCIIFrame) SetData(data []byte) {
	frame.Data = data
}

// SetException sets the Modbus exception code in the frame.
func (frame *ASCIIFrame) SetException(exception *Exception) {
	frame.Function = frame.Function | 0x80
	frame.Data = []byte{byte(*exception)}
}
