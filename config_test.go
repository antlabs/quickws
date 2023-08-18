package quickws

import "testing"

func Test_InitPayloadSize(t *testing.T) {
	t.Run("InitPayload", func(t *testing.T) {
		var c Config
		for i := 1; i < 32; i++ {
			c.windowsMultipleTimesPayloadSize = float32(i)
			if c.initPayloadSize() != i*(1024+14) {
				t.Errorf("initPayloadSize() = %d, want %d", c.initPayloadSize(), i*(1024+14))
			}
		}
	})
}
