# BENZHI 评测说明

基于 Go 实现的量子纠错综合征时空关联后端服务，一款后端服务，完成稳定子测量综合征入库、码格邻接上的时空缺陷图重建、短暂噪声与跨轮次持续错误链判定，并发布不可变解码证据快照。

## 启动

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/qecorr --addr :8080 --db qecorr.db
```

## 自检（不启动长驻服务）

```bash
go run ./cmd/qecorr --smoke-test
```

`--smoke-test` 会真实创建码格与测量轮次、写入综合征、重建缺陷图、评分错误链并发布快照，关闭并重新打开数据库验证持久化与重启恢复，最后以 0 退出码结束。

## 构建门禁

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/qecorr --smoke-test
```

## HTTP API（前缀 /api）

码格：`POST/GET /api/lattices`、`GET /api/lattices/{id}`
轮次：`POST/GET /api/rounds`、`GET /api/rounds/{id}`
量子比特：`/api/qubits/{id}`
错误链：`/api/chains/{id}`
快照：`/api/snapshots/{id}`
自检：`GET /api/health`、`GET /api/selfcheck`

## 持久化

SQLite（modernc.org/sqlite，CGO 无关）。建表：lattices、qubits、adjacency、rounds、syndromes、calibrations、defect_edges、error_chains、snapshots。轮次号与综合征键由数据库约束保护。
