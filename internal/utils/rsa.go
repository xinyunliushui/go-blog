/*
 * @Date: 2026-03-31 17:14:29
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-31 17:56:26
 * @Description: rsa encryption and decryption  HS256 适合单体；RS256 适合微服务 / 分布式
 */
package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

/** 从文件中读取RSA key
 * @param filename string 文件名
 * @return []byte RSA key
 */
func RSAReadKeyFromFile(filename string) []byte {
	f, err := os.Open(filename)
	var b []byte

	if err != nil {
		return b
	}
	defer f.Close()
	fileInfo, _ := f.Stat()
	b = make([]byte, fileInfo.Size())
	f.Read(b)
	return b
}

/** RSA加密
 * @param data []byte 数据
 * @param publicBytes []byte RSA公钥
 * @return []byte 加密后的数据
 * @return error 错误
 */
func RSAEncrypt(data, publicBytes []byte) ([]byte, error) {
	var res []byte
	// 解析公钥
	block, _ := pem.Decode(publicBytes)

	if block == nil {
		return res, fmt.Errorf("无法加密, 公钥可能不正确")
	}

	// 使用X509将解码之后的数据 解析出来
	// x509.MarshalPKCS1PublicKey(block):解析之后无法用，所以采用以下方法：ParsePKIXPublicKey
	keyInit, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return res, fmt.Errorf("无法加密, 公钥可能不正确, %v", err)
	}
	// 使用公钥加密数据
	pubKey := keyInit.(*rsa.PublicKey)
	res, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		return res, fmt.Errorf("无法加密, 公钥可能不正确, %v", err)
	}
	// 将数据加密为base64格式
	return []byte(EncodeStr2Base64(string(res))), nil
}

/** 对数据进行解密操作
 * @param base64Data []byte base64数据
 * @param privateBytes []byte RSA私钥
 * @return []byte 解密后的数据
 * @return error 错误
 */
func RSADecrypt(base64Data, privateBytes []byte) ([]byte, error) {
	var res []byte
	// 将base64数据解析
	data := []byte(DecodeStrFromBase64(string(base64Data)))
	// 解析私钥
	block, _ := pem.Decode(privateBytes)
	if block == nil {
		return res, fmt.Errorf("无法解密, 私钥可能不正确")
	}
	// 还原数据
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return res, fmt.Errorf("无法解密, 私钥可能不正确, %v", err)
	}
	res, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return res, fmt.Errorf("无法解密, 私钥可能不正确, %v", err)
	}
	return res, nil
}

/** 加密base64字符串
 * @param str string 字符串
 * @return string base64字符串
 */
func EncodeStr2Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

/** 解密base64字符串
 * @param str string base64字符串
 * @return string 解密后的字符串
 */
func DecodeStrFromBase64(str string) string {
	decodeBytes, _ := base64.StdEncoding.DecodeString(str)
	return string(decodeBytes)
}
