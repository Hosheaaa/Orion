package domain

import "errors"

// 领域错误定义
var (
	// ErrActivityNotFound 活动不存在
	ErrActivityNotFound = errors.New("活动不存在")
	// ErrActivityAlreadyExists 活动已存在
	ErrActivityAlreadyExists = errors.New("活动已存在")
	// ErrInvalidActivityStatus 无效的活动状态
	ErrInvalidActivityStatus = errors.New("无效的活动状态")
	// ErrActivityCannotBeModified 活动无法修改
	ErrActivityCannotBeModified = errors.New("活动无法修改")
)
