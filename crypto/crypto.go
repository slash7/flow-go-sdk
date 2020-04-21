package crypto

import (
	"encoding/hex"

	"github.com/onflow/flow-go-sdk/crypto/internal/crypto"
)

// SignatureAlgorithm is an identifier for a signature algorithm (and parameters if applicable).
type SignatureAlgorithm int

const (
	UnknownSignatureAlgorithm SignatureAlgorithm = iota
	// BLS_BLS12381 is BLS on BLS 12-381 curve
	BLS_BLS12381
	// ECDSA_P256 is ECDSA on NIST P-256 curve
	ECDSA_P256
	// ECDSA_secp256k1 is ECDSA on secp256k1 curve
	ECDSA_secp256k1
)

// String returns the string representation of this signature algorithm.
func (f SignatureAlgorithm) String() string {
	return [...]string{"UNKNOWN", "BLS_BLS12381", "ECDSA_P256", "ECDSA_secp256k1"}[f]
}

// StringToSignatureAlgorithm converts a string to a SignatureAlgorithm.
func StringToSignatureAlgorithm(s string) SignatureAlgorithm {
	switch s {
	case BLS_BLS12381.String():
		return BLS_BLS12381
	case ECDSA_P256.String():
		return ECDSA_P256
	case ECDSA_secp256k1.String():
		return ECDSA_secp256k1
	default:
		return UnknownSignatureAlgorithm
	}
}

// HashAlgorithm is an identifier for a hash algorithm.
type HashAlgorithm int

const (
	UnknownHashAlgorithm HashAlgorithm = iota
	SHA2_256
	SHA2_384
	SHA3_256
	SHA3_384
)

// String returns the string representation of this hash algorithm.
func (f HashAlgorithm) String() string {
	return [...]string{"UNKNOWN", "SHA2_256", "SHA2_384", "SHA3_256", "SHA3_384"}[f]
}

// StringToHashAlgorithm converts a string to a HashAlgorithm.
func StringToHashAlgorithm(s string) HashAlgorithm {
	switch s {
	case SHA2_256.String():
		return SHA2_256
	case SHA2_384.String():
		return SHA2_384
	case SHA3_256.String():
		return SHA3_256
	case SHA3_384.String():
		return SHA3_384
	default:
		return UnknownHashAlgorithm
	}
}

const (
	MinSeedLengthECDSA_P256      = crypto.KeyGenSeedMinLenECDSAP256
	MinSeedLengthECDSA_secp256k1 = crypto.KeyGenSeedMinLenECDSASecp256k1
)

// KeyType is a key format supported by Flow.
type KeyType int

const (
	UnknownKeyType KeyType = iota
	ECDSA_P256_SHA2_256
	ECDSA_P256_SHA3_256
	ECDSA_secp256k1_SHA2_256
	ECDSA_secp256k1_SHA3_256
)

// SignatureAlgorithm returns the signature algorithm for this key type.
func (k KeyType) SignatureAlgorithm() SignatureAlgorithm {
	switch k {
	case ECDSA_P256_SHA2_256, ECDSA_P256_SHA3_256:
		return ECDSA_P256
	case ECDSA_secp256k1_SHA2_256, ECDSA_secp256k1_SHA3_256:
		return ECDSA_secp256k1
	default:
		return UnknownSignatureAlgorithm
	}
}

// HashAlgorithm returns the hash algorithm for this key type.
func (k KeyType) HashAlgorithm() HashAlgorithm {
	switch k {
	case ECDSA_P256_SHA2_256, ECDSA_secp256k1_SHA2_256:
		return SHA2_256
	case ECDSA_P256_SHA3_256, ECDSA_secp256k1_SHA3_256:
		return SHA3_256
	default:
		return UnknownHashAlgorithm
	}
}

// A PrivateKey is a cryptographic private key that can be used for in-memory signing.
type PrivateKey struct {
	private crypto.PrivateKey
}

// Sign signs the given message with this private key and the provided hasher.
//
// This function returns an error if a signature cannot be generated.
func (pk PrivateKey) Sign(message []byte, hasher Hasher) ([]byte, error) {
	return pk.private.Sign(message, hasher)
}

// Algorithm returns the signature algorithm for this private key.
func (pk PrivateKey) Algorithm() SignatureAlgorithm {
	return SignatureAlgorithm(pk.private.Algorithm())
}

// PublicKey returns the public key for this private key.
func (pk PrivateKey) PublicKey() PublicKey {
	return PublicKey{publicKey: pk.private.PublicKey()}
}

// Encode returns the raw byte encoding of this private key.
func (pk PrivateKey) Encode() []byte {
	return pk.private.Encode()
}

// A PublicKey is a cryptographic public key that can be used to verify signatures.
type PublicKey struct {
	publicKey crypto.PublicKey
}

// Verify verifies the given signature against a message with this public key and the provided hasher.
//
// This function returns true if the signature is valid for the message, and false otherwise. An error
// is returned if the signature cannot be verified.
func (pk PublicKey) Verify(sig, message []byte, hasher Hasher) (bool, error) {
	return pk.publicKey.Verify(sig, message, hasher)
}

// Algorithm returns the signature algorithm for this public key.
func (pk PublicKey) Algorithm() SignatureAlgorithm {
	return SignatureAlgorithm(pk.publicKey.Algorithm())
}

// Encode returns the raw byte encoding of this public key.
func (pk PublicKey) Encode() []byte {
	return pk.publicKey.Encode()
}

// A Signer is capable of signing cryptographic messages.
type Signer interface {
	// Sign signs the given message with this signer.
	Sign(message []byte) ([]byte, error)
}

// An InMemorySigner is a signer that generates signatures using an in-memory private key.
type InMemorySigner struct {
	PrivateKey PrivateKey
	Hasher     Hasher
}

// NewInMemorySigner initializes and returns a new in-memory signer with the provided private key
// and hasher.
func NewInMemorySigner(privateKey PrivateKey, hashAlgo HashAlgorithm) InMemorySigner {
	hasher, _ := NewHasher(hashAlgo)

	return InMemorySigner{
		PrivateKey: privateKey,
		Hasher:     hasher,
	}
}

func (s InMemorySigner) Sign(message []byte) ([]byte, error) {
	return s.PrivateKey.Sign(message, s.Hasher)
}

// NaiveSigner is an alias for InMemorySigner.
type NaiveSigner = InMemorySigner

// NewNaiveSigner is an alias for NewInMemorySigner.
func NewNaiveSigner(privateKey PrivateKey, hashAlgo HashAlgorithm) NaiveSigner {
	return NewInMemorySigner(privateKey, hashAlgo)
}

// GeneratePrivateKey generates a private key with the specified signature algorithm from the given seed.
func GeneratePrivateKey(sigAlgo SignatureAlgorithm, seed []byte) (PrivateKey, error) {
	hasher := NewSHA3_384()
	hashedSeed := hasher.ComputeHash(seed)

	privKey, err := crypto.GeneratePrivateKey(crypto.SigningAlgorithm(sigAlgo), hashedSeed)
	if err != nil {
		return PrivateKey{}, err
	}

	return PrivateKey{
		private: privKey,
	}, nil
}

// DecodePrivateKey decodes a raw byte encoded private key with the given signature algorithm.
func DecodePrivateKey(sigAlgo SignatureAlgorithm, b []byte) (PrivateKey, error) {
	privKey, err := crypto.DecodePrivateKey(crypto.SigningAlgorithm(sigAlgo), b)
	if err != nil {
		return PrivateKey{}, err
	}

	return PrivateKey{
		private: privKey,
	}, nil
}

// DecodePrivateKeyHex decodes a raw hex encoded private key with the given signature algorithm.
func DecodePrivateKeyHex(sigAlgo SignatureAlgorithm, s string) (PrivateKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return PrivateKey{}, err
	}

	return DecodePrivateKey(sigAlgo, b)
}

// DecodePublicKey decodes a raw byte encoded public key with the given signature algorithm.
func DecodePublicKey(sigAlgo SignatureAlgorithm, b []byte) (PublicKey, error) {
	pubKey, err := crypto.DecodePublicKey(crypto.SigningAlgorithm(sigAlgo), b)
	if err != nil {
		return PublicKey{}, err
	}

	return PublicKey{
		publicKey: pubKey,
	}, nil
}

// DecodePublicKeyHex decodes a raw hex encoded public key with the given signature algorithm.
func DecodePublicKeyHex(sigAlgo SignatureAlgorithm, s string) (PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return PublicKey{}, err
	}

	return DecodePublicKey(sigAlgo, b)
}

// CompatibleAlgorithms returns true if the signature and hash algorithms are compatible.
func CompatibleAlgorithms(sigAlgo SignatureAlgorithm, hashAlgo HashAlgorithm) bool {
	switch sigAlgo {
	case ECDSA_P256:
		fallthrough
	case ECDSA_secp256k1:
		switch hashAlgo {
		case SHA2_256:
			fallthrough
		case SHA3_256:
			return true
		}
	}
	return false
}
