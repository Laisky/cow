package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"testing"
	"time"

	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

// These fixtures were independently produced by both reviewed module graphs.
// All passwords, IVs, addresses and bytes here are synthetic test values.
const compatibilityPassword = "cow-compatibility-only"

func compatibilityDecode(t *testing.T, method, password string, wire []byte) ([]byte, error) {
	t.Helper()
	proxy := newCowProxy(method, password, "127.0.0.1:0")
	a, b := net.Pipe()
	deadline := time.Now().Add(5 * time.Second)
	if err := a.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	if err := b.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	wrapped := ss.NewConn(a, proxy.cipher.Copy())
	defer wrapped.Close()
	done := make(chan error, 1)
	go func() {
		defer b.Close()
		// Fragment IV and payload across reads, including Salsa20 block boundaries.
		for len(wire) > 0 {
			n := 7
			if len(wire) < n {
				n = len(wire)
			}
			if _, err := b.Write(wire[:n]); err != nil {
				done <- err
				return
			}
			wire = wire[n:]
		}
		done <- nil
	}()
	plain, err := io.ReadAll(wrapped)
	// On truncated IV, the writer has closed; always join its bounded goroutine.
	if writerErr := <-done; writerErr != nil {
		t.Fatal(writerErr)
	}
	return plain, err
}

func TestCryptoDependencyWireCompatibility(t *testing.T) {
	raw, err := os.ReadFile("testdata/crypto-wire.json")
	if err != nil {
		t.Fatal(err)
	}
	fixtures := map[string]string{}
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	plain, err := hex.DecodeString(fixtures["plaintext"])
	if err != nil || len(plain) == 0 {
		t.Fatal("invalid plaintext fixture", err)
	}
	for _, method := range []string{"aes-128-cfb", "bf-cfb", "cast5-cfb", "salsa20"} {
		t.Run(method, func(t *testing.T) {
			wire, err := hex.DecodeString(fixtures[method])
			if err != nil || len(wire) <= len(plain) {
				t.Fatal("invalid wire fixture", err)
			}
			t.Run("recorded_wire", func(t *testing.T) {
				got, err := compatibilityDecode(t, method, compatibilityPassword, wire)
				if err != nil || !bytes.Equal(got, plain) {
					t.Fatalf("fixed wire mismatch: %v", err)
				}
			})
			t.Run("wrong_password_control", func(t *testing.T) {
				got, err := compatibilityDecode(t, method, "different-synthetic-password", wire)
				if err != nil {
					t.Fatal(err)
				}
				// Legacy stream ciphers do not authenticate: test inequality, not rejection.
				if bytes.Equal(got, plain) {
					t.Fatal("wrong password falsely matched")
				}
			})
			t.Run("truncated_iv", func(t *testing.T) {
				got, err := compatibilityDecode(t, method, compatibilityPassword, wire[:3])
				if !errors.Is(err, io.ErrUnexpectedEOF) || len(got) != 0 {
					t.Fatalf("truncated IV: bytes=%d err=%v", len(got), err)
				}
			})
			t.Run("changed_ciphertext_control", func(t *testing.T) {
				changed := append([]byte(nil), wire...)
				changed[len(changed)-1] ^= 1
				got, err := compatibilityDecode(t, method, compatibilityPassword, changed)
				if err != nil {
					t.Fatal(err)
				}
				if bytes.Equal(got, plain) {
					t.Fatal("changed ciphertext falsely matched")
				}
			})
			t.Run("parent_to_listener", func(t *testing.T) {
				parent := newShadowsocksParent("127.0.0.1:0")
				parent.initCipher(method, compatibilityPassword)
				listener := newCowProxy(method, compatibilityPassword, "127.0.0.1:0")
				a, b := net.Pipe()
				deadline := time.Now().Add(5 * time.Second)
				if err := a.SetDeadline(deadline); err != nil {
					t.Fatal(err)
				}
				if err := b.SetDeadline(deadline); err != nil {
					t.Fatal(err)
				}
				sender := ss.NewConn(a, parent.cipher.Copy())
				receiver := ss.NewConn(b, listener.cipher.Copy())
				defer receiver.Close()
				done := make(chan error, 1)
				go func() {
					defer sender.Close()
					for _, part := range [][]byte{plain[:1], plain[1:64], plain[64:65], plain[65:]} {
						// Upstream's first Write count includes its IV; only its error is relevant.
						if _, err := sender.Write(part); err != nil {
							done <- err
							return
						}
					}
					done <- nil
				}()
				got, err := io.ReadAll(receiver)
				if sendErr := <-done; sendErr != nil {
					t.Fatal(sendErr)
				}
				if err != nil || !bytes.Equal(got, plain) {
					t.Fatalf("parent/listener mismatch: %v", err)
				}
			})
		})
	}
}

func TestCryptoDependencyInvalidConfiguration(t *testing.T) {
	if _, err := ss.NewCipher("unsupported-compatibility-test", compatibilityPassword); err == nil {
		t.Fatal("unsupported cipher accepted")
	}
	if _, err := ss.NewCipher("aes-128-cfb", ""); err == nil {
		t.Fatal("empty password accepted")
	}
}
