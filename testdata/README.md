# Crypto dependency compatibility fixtures

`crypto-wire.json` contains synthetic binary plaintext and fixed-IV Shadowsocks
packets. The same standalone producer, pinned shadowsocks-go revision
`3e585ff90601`, password `cow-compatibility-only`, and test-only IV bytes `0x42`
produced byte-identical packets under both COW module graphs:

- original `21aeaeb7e8ea2d612582083a52154f0abf81d1b2`, x/crypto v0.17.0;
- dependency PR original head `b6a90a64a3687e9af38c7625774b8878ed8b1e70`, x/crypto v0.52.0.

The persistent tests decode these fixed packets through the real COW listener
constructor, exercise fragmented binary traffic from the COW parent constructor,
and check truncated IV, wrong-password, modified-ciphertext and invalid-config
controls. AES is a standard-library reference; Blowfish, CAST5 and Salsa20 are
the actual x/crypto import paths used by the pinned Shadowsocks dependency.
The upstream package known-answer tests run separately under both graphs.
No network listener, home proxy configuration, real key or public website is used.

Fixed IVs and this public synthetic password are **fixtures only**. No production
cipher, key derivation, authentication or configuration behavior is changed.
These legacy streams are unauthenticated; wrong passwords or modified ciphertext
produce different plaintext, not guaranteed authentication errors.

The read-only compatibility workflow runs Go 1.25.8 and 1.27.2 against the
immutable original graph and candidate, verifies module-file immutability,
race tests, vet, amd64 build and arm64 cross-compilation. Cross-compilation is
not native arm64 runtime acceptance. The dependency update raises the declared
minimum Go version from 1.17 to 1.25; these checks do not promise old-Go support.
The inherited external-site shell integration test and mutating Makefile tool
installation/lint targets are separate and are not invoked by this workflow.
