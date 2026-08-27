# 量子纠错综合征时空关联服务（task245-qecorr）

```bash
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go run ./cmd/qecorr --smoke-test --db /tmp/qecorr.db
```

运行服务：

```bash
go run ./cmd/qecorr --addr :8080 --db qecorr.db
```

服务接收码格、轮次和综合征数据，重建缺陷图与错误链，并将证据发布为不可变快照。
