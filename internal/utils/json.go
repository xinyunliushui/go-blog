/*
 * @Date: 2026-03-31 17:14:29
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-24 11:00:16
 * @Description: json工具
 */
package utils

import (
	"encoding/json"
	"fmt"
)

/** 结构体转为json
 * @param obj interface{} 结构体
 * @return string json字符串
 */
func Struct2Json(obj interface{}) string {
	str, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("[Struct2Json]转换异常: %v", err))
	}
	return string(str)
}

/** json转为结构体
 * @param str string json字符串
 * @param obj interface{} 结构体
 * @return void
 */
func Json2Struct(str string, obj interface{}) {
	// 将json转为结构体
	err := json.Unmarshal([]byte(str), obj)
	if err != nil {
		panic(fmt.Sprintf("[Json2Struct]转换异常: %v", err))
	}
}

/** json interface转为结构体
 * @param str interface{} json字符串
 * @param obj interface{} 结构体
 * @return void
 */
func JsonI2Struct(str interface{}, obj interface{}) {
	JsonStr := str.(string)
	Json2Struct(JsonStr, obj)
}
