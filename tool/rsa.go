package tool

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// 加密
func RsaEncrypt(origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

// 公钥和私钥可以从文件中读取

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQC3K/puBN2P92uD7SpoSL92SNYs0r8Q26eKHjJwxBx0f4Hlvul1
uBZg4ZkE3z1B5rPmFWtaGBa227gUc1LbE5NWrEipa0VSP4qGkFHNaL3fRzqMc2K8
JAUEKMpFrkhcw70CN0bWbzFvWhex2JbLuy71c5cmnV8uvtsZrgVgPSfc5wIDAQAB
AoGAN+sWFZYoqXWn/eteId3rjUmpEJ/5skTMPc8AKQrFgQ8X8bI5hTWAp2zXkPQx
uDecveXWEvf9ny8uYBfguH6eYLgS6W4TQFfdgdbMeRFH8cC7qVYD686TmYgNj85i
1ua+76JOwCBsPg0p5OgLyDJAIjA7PS1K2XaWpwJVHhAv80kCQQDdNGAp12nP1DE4
j3kweJ5eBsmpfy31Vg9A7jBp1+7P4O38tfHxOtXSnX1jwqscFM1KGhPe8vQpBacl
v/QSEN2TAkEA0/wSEZQB9kJR9fwJ2vpsxzL+BfKhfAOsp2HCeyG7vep00a0Z4rKc
bbh+G6+rfanLNfIo6WPJdFiu8/Todco33QJBALrm3DG+TytJMOWHZHBuGfF8brwG
N4DJ3E2Sc9mal6+Rb8RMv0aB3dT9OMsn2of5k5N/ATcptN9MZXRiAgmZsn0CQQCT
RGXlEla+ltpLsnnCSAEz7efth97Jwd+7NL4gPpIn4O6hD8mQ5RapXuc1IrhXh5Ll
+kKTyUAV9NouHvEzi3V9AkEAyAu/toXI5/zjojIotMRaTV6k2ayKkEmJWVSpjKCt
6SNa5z1rCDkdboPPFfKqRbIeSIIRTRimlEXF81XZnBjvXA==
-----END RSA PRIVATE KEY-----
`)

var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDENd/iIkAzO6srPpK5VmMbjnz/
PXfepLegBqHv/Wr3vURgv20pqH+1OF6puvtnW9MsWVBYMKUDC9w/YvqMMkDyS7lM
oZm1/xwJRhDpxjmMVte3SAs2jqOLVfetr1MYzn6zdBX0kR4tiUYFgpkrNsRoM1Jh
03UnJMOEB8srQdIWywIDAQAB
-----END PUBLIC KEY-----
`)
