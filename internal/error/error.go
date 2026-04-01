/*
 * @Date: 2026-03-31 17:21:16
 * @Author: zhongwenhao
 * @LastEditors: zhongwenhao
 * @LastEditTime: 2026-03-31 17:22:01
 * @Description: standardized error
 */
package error

import "errors"

var (
	ErrUserNotFound = errors.New("resource not found")
	ErrUserExists   = errors.New("resource already exists")
	ErrUserDisabled = errors.New("resource disabled")
	ErrUserLocked   = errors.New("resource locked")
	ErrUserExpired  = errors.New("resource expired")
	ErrUserInvalid  = errors.New("resource invalid")
)
