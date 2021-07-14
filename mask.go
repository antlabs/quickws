package tinyws

func mask(payload []byte, maskVal []byte) {
	for i := 0; i < len(payload); i++ {
		payload[i] ^= maskVal[i%4]
	}
}
