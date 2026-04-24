/*
 * @Date: 2026-03-31 17:14:29
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 10:58:39
 * @Description: bcrypt密码加密
 */
package utils

import "golang.org/x/crypto/bcrypt"

/** 密码加密 使用自适应hash算法, 不可逆
 * @param passwd string 明文
 * @return string 密文
 */
func GenPasswd(passwd string) string {
	hashPasswd, _ := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	return string(hashPasswd)
}

/** 通过比较两个字符串hash判断是否出自同一个明文
 * @param hashPasswd string 需要对比的密文
 * @param passwd string 明文
 * @return error
 */
func ComparePasswd(hashPasswd string, passwd string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPasswd), []byte(passwd)); err != nil {
		return err
	}
	return nil
}
