package quickws

import "testing"

func Test_FixedWriter(t *testing.T) {
	fw := &fixedWriter{buf: make([]byte, 1024)}
	n, err := fw.Write([]byte("hello"))
	if err != nil {
		t.Errorf("fw.Write() = %v, want nil", err)
	}
	if n != 5 {
		t.Errorf("fw.Write() = %d, want 5", n)
	}

	if fw.Len() != 5 {
		t.Errorf("fw.Len() = %d, want 5", fw.Len())
	}
	if string(fw.Bytes()) != "hello" {
		t.Errorf("fw.Bytes() = %s, want hello", fw.Bytes())
	}
}
