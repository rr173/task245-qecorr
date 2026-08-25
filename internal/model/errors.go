// Package model 定义量子纠错综合征时空关联服务的领域实体、状态枚举与公共错误。
package model

import "errors"

// 领域错误集合，供各业务包与存储层统一返回。
var (
	// ErrNotFound 实体不存在。
	ErrNotFound = errors.New("not found")
	// ErrDuplicate 唯一键冲突（幂等入口重复写入）。
	ErrDuplicate = errors.New("duplicate")
	// ErrInvalidState 状态流转非法。
	ErrInvalidState = errors.New("invalid state transition")
	// ErrUnknownQubit 量子比特不属于该码格或不存在。
	ErrUnknownQubit = errors.New("unknown qubit")
	// ErrRoundRegression 轮次序号倒退（拒绝旧轮次回灌）。
	ErrRoundRegression = errors.New("round regression")
	// ErrAsymmetricAdjacency 邻接边不对称（a-b 与 b-a 必须同时声明）。
	ErrAsymmetricAdjacency = errors.New("asymmetric adjacency")
	// ErrSealed 实体已封存，拒绝后续写入。
	ErrSealed = errors.New("sealed, modification rejected")
	// ErrBadRequest 入参非法。
	ErrBadRequest = errors.New("bad request")
)
