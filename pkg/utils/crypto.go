// pkg/utils/crypto.go
package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"
)

// ------------------------------------------------------------------
// 对称加密 (AES-GCM)
// ------------------------------------------------------------------

// NewAESKey 随机 32 字节 key
func NewAESKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// AESEncrypt AES-GCM 加密后 base64 output
func AESEncrypt(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipher := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(cipher), nil
}

// AESDecrypt AES-GCM 解码
func AESDecrypt(enc string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return nil, err
	}
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("cipher too short")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ------------------------------------------------------------------
// 非对称加密 (RSA)
// ------------------------------------------------------------------

type RSAKeyPair struct {
	PrivatePEM string `json:"private"`
	PublicPEM  string `json:"public"`
}

// GenerateRSAKeyPair 2048bit
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}
	return &RSAKeyPair{
		PrivatePEM: string(pemEncode(privBytes, "RSA PRIVATE KEY")),
		PublicPEM:  string(pemEncode(pubBytes, "PUBLIC KEY")),
	}, nil
}

// RSAEncrypt base64
func RSAEncrypt(msg []byte, pubPEM string) (string, error) {
	pubBlock, _ := pem.Decode([]byte(pubPEM))
	if pubBlock == nil {
		return "", errors.New("invalid PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return "", err
	}
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), msg)
	return base64.StdEncoding.EncodeToString(encrypted), err
}

// RSADecrypt
func RSADecrypt(enc string, privPEM string) ([]byte, error) {
	privBlock, _ := pem.Decode([]byte(privPEM))
	if privBlock == nil {
		return nil, errors.New("invalid PEM")
	}
	priv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}
	data, _ := base64.StdEncoding.DecodeString(enc)
	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
}

func pemEncode(b []byte, t string) []byte {
	block := &pem.Block{Type: t, Bytes: b}
	return pem.EncodeToMemory(block)
}

// ------------------------------------------------------------------
// 哈希
// ------------------------------------------------------------------

// SHA256 base64
func SHA256(data []byte) string {
	h := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(h[:])
}

// ------------------------------------------------------------------
// 安全随机字符串 & UUID
// ------------------------------------------------------------------

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._"

// RandomString n 字节随机符号
func RandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = letters[v%byte(len(letters))]
	}
	return string(b), nil
}

// SecureToken 32字节 URL-safe token
func SecureToken() (string, error) {
	raw, err := RandomString(32)
	return SHA256([]byte(raw))[:32], err // truncate for compact
}

// ------------------------------------------------------------------
// 轻量级 JWT 实现（HS256）
// ------------------------------------------------------------------

type Claims map[string]interface{}

// SignHS256 签名； key == []byte
func SignHS256(claims Claims, secret []byte) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{\"alg\":\"HS256\",\"typ\":\"JWT\"}`))
	payload, _ := json.Marshal(claims)
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := sha256.Sum256([]byte(header + "." + body))
	return header + "." + body + "." + base64.RawURLEncoding.EncodeToString(mac[:]), nil
}

// VerifyHS256 解析 & 校验签名，返回 Claims
func VerifyHS256(token string, secret []byte) (Claims, error) {
	parts := bytes.Split([]byte(token), []byte{'.'})
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT")
	}
	mac := sha256.Sum256([]byte(parts[0])) // 再次签名
	expected := base64.RawURLEncoding.EncodeToString(mac[:])
	if string(parts[2]) != expected {
		return nil, errors.New("invalid signature")
	}
	var c Claims
	body, _ := base64.RawURLEncoding.DecodeString(string(parts[1]))
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, err
	}
	return c, nil
}

// ------------------------------------------------------------------
// KeyRotation 内存轻量轮换示例
// ------------------------------------------------------------------

type KeyRotator struct {
	key   []byte // current key
	gen   time.Time
	valid time.Duration
}

// NewKeyRotator valid 默认 24h
func NewKeyRotator(valid time.Duration) *KeyRotator {
	k, _ := NewAESKey()
	return &KeyRotator{key: k, gen: time.Now(), valid: valid}
}

// Valid return ok=true if still valid
func (r *KeyRotator) Valid() (key []byte, ok bool) {
	if time.Since(r.gen) > r.valid {
		return nil, false
	}
	return r.key, true
}

// Rotate 手动立即轮换
func (r *KeyRotator) Rotate() error {
	k, err := NewAESKey()
	if err != nil {
		return err
	}
	r.key, r.gen = k, time.Now()
	return nil
}

// ------------------------------------------------------------------
// EC 自签名证书生成（TLS 开发用）
// ------------------------------------------------------------------

// GenSelfSigned 返回 PEM 证书和私钥
func GenSelfSigned(host string, validDays int) (certPEM, keyPEM string, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(validDays) * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("agentic-ai-%s", host),
			Organization: []string{"AgenticAI"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		DNSNames:              []string{host},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derCert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, priv.Public(), priv)
	if err != nil {
		return "", "", err
	}
	certPEM = string(pemEncode(derCert, "CERTIFICATE"))

	privPKCS8, _ := x509.MarshalECPrivateKey(priv)
	keyPEM = string(pemEncode(privPKCS8, "EC PRIVATE KEY"))
	return
}
//Personal.AI order the ending
