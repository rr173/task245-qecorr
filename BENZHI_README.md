# 构建说明

本项目是纯 Go 后端服务，使用 `modernc.org/sqlite`，无需 CGO 或外部数据库。构建入口为 `cmd/qecorr`；`--smoke-test` 会在临时 SQLite 文件上执行写入、关联、快照发布和重启恢复自检。
