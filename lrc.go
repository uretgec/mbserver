package mbserver

// Check Hash Methods
/*
via: https://github.com/things-go/go-modbus/blob/master/lrc.go
LRC

The Longitudinal Redundancy Check (LRC) field is one byte, containing an 8–bit
binary value. The LRC value is calculated by the transmitting device, which
appends the LRC to the message. The receiving device recalculates an LRC
during receipt of the message, and compares the calculated value to the actual
value it received in the LRC field. If the two values are not equal, an error results.
*/
func lrcModbus(data ...byte) (sum uint8) {
	if len(data) > 0 {
		for _, b := range data {
			sum += b
		}
	}

	return uint8(-int8(sum))
}
