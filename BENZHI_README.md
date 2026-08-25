基于 Go 实现的 LaunchGuard 项目，一款火箭发射场安全联控服务，协调推进剂加注、脐带撤收和倒计时许可。

服务默认监听 `127.0.0.1:21220`，运行 `go run -mod=vendor ./cmd/launchguard` 后可访问四个运营页面和 JSON API。

本地状态保存在 `runtime` 目录。项目使用 vendor 目录离线构建，可执行 `go build -mod=vendor ./...`。
