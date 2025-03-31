package main

// returns true and a respective status code if the connection should be closed
// and false if should not close
func opcodeStatusCheck(opcode Opcode, status StatusCode) (bool, StatusCode) {
	if opcode == OpcodeConnectionClose {
		if status <= 999 ||
			status == 1002 ||
			status == 1004 ||
			status == 1005 ||
			status == 1006 ||
			(status >= 1016 && status < 3000) {
			return true, StatusProtocolError
		}

		return true, status
		// return true, StatusNormalClosure
	} else if (opcode >= 0x3 && opcode <= 0x7) || opcode >= 0xB {
		// * %x3-7 are reserved for further non-control frames
		// * %xB-F are reserved for further control frames
		return true, StatusProtocolError
	}
	return false, 0
}
