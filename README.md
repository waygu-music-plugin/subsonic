# Subsonic 兼容平台（Navidrome、Jellyfin、Gonic、通用 Subsonic）

用于 [Wagyu Music](https://github.com/vicnoah/mygo-cloudmusic) 的外置平台插件，支持 Windows、macOS 和 Linux 的 amd64、arm64。

在宿主的「插件管理」中选择此插件包，从[官方目录](https://github.com/waygu-music-plugin/registry)安装。

在当前目录运行 `go build ./cmd/subsonic-plugin` 可构建当前平台的插件程序。六个目标的 ZIP 和签名清单见本仓库的 Releases。

使用 [插件 SDK](https://github.com/waygu-music-plugin/sdk) 的 v1 通信协议。运行 `go test ./...` 可执行本仓库的测试。
