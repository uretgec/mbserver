package mbserver

import (
	"encoding/binary"
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
	dataEnd := pLen - 2
	lrcExpect := packet[dataEnd:]
	lrcCalc := lrcModbus(packet[1:dataEnd]...)
	if lrcExpect[1] != lrcCalc {
		err := fmt.Errorf("modbus: response lrc '%v' does not match expected '%v'", lrcCalc, lrcExpect)
		return nil, err
	}

	frame := &ASCIIFrame{
		Address:  uint8(packet[1]),
		Function: uint8(packet[2]),
		Data:     packet[2 : pLen-2],
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
	bytes := make([]byte, 2)

	bytes[0] = frame.Address
	bytes[1] = frame.Function
	bytes = append(bytes, frame.Data...)

	// Calculate the CRC.
	pLen := len(bytes)
	crc := crcModbus(bytes[0:pLen])

	// Add the CRC.
	bytes = append(bytes, []byte{0, 0}...)
	binary.LittleEndian.PutUint16(bytes[pLen:pLen+2], crc)

	return bytes
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
