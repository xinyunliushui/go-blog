/*
 * @Date: 2026-03-27 21:51:44
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-04-30 10:14:51
 * @Description: validator
 */
package common

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	ch_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 全局Validate数据校验实例
var Validate = validator.New()

// 全局翻译器
var Trans ut.Translator

/** 注册标签函数
 * @param v *validator.Validate 验证器
 * @return void
 */
func registerTagNameFunc(v *validator.Validate) {
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		for _, tag := range []string{"json", "form", "uri"} {
			if s := fld.Tag.Get(tag); s != "" && s != "-" {
				i := strings.Index(s, ",")
				if i > 0 {
					return s[:i]
				}
				return s
			}
		}
		return fld.Name
	})
}

/** 将 binding / validator 错误转为中文提示（校验类错误走翻译，其余给常用中文说明）。
 * @param err error 错误
 * @return string 错误提示
 */
func ValidationErrString(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) && len(ve) > 0 {
		return ve[0].Translate(Trans)
	}
	var syn *json.SyntaxError
	if errors.As(err, &syn) {
		return "请求参数JSON格式错误"
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return "请求参数类型不正确"
	}
	return err.Error()
}

// 初始化Validator数据校验
func InitValidator() {
	chinese := zh.New()
	uni := ut.New(chinese, chinese)
	trans, _ := uni.GetTranslator("zh")
	Trans = trans

	// Validate = validator.New()
	// registerTagNameFunc(Validate)
	// _ = ch_translations.RegisterDefaultTranslations(Validate, Trans)

	// 注册手机号校验
	_ = Validate.RegisterValidation("checkMobile", checkMobile)

	// Gin ShouldBind/ShouldBindJSON 等使用 binding 内置引擎，需单独注册翻译与 TagName
	if gv, ok := binding.Validator.Engine().(*validator.Validate); ok {
		registerTagNameFunc(gv)
		_ = ch_translations.RegisterDefaultTranslations(gv, Trans)
	}
}

/** 手机号校验
 * @param fl validator.FieldLevel 字段级别
 * @return bool 是否校验通过
 */
func checkMobile(fl validator.FieldLevel) bool {
	reg := `^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(fl.Field().String())
}
