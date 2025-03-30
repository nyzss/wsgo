package main

// returns true and a respective status code if the connection should be closed
// and false if should not close
func opcodeCheck(opcode Opcode) (bool, StatusCode) {
	if opcode == OpcodeConnectionClose {
		return true, StatusNormalClosure
	} else if (opcode >= 0x3 && opcode <= 0x7) || opcode >= 0xB {
		// * %x3-7 are reserved for further non-control frames
		// * %xB-F are reserved for further control frames
		return true, StatusProtocolError
	}
	return false, 0
}
