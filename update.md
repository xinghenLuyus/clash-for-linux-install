# re-dev-go 变更记录

## Go TUI 第一版

- 从 `master` 基线重新开始，不继承旧 `dev-go` 实现。
- 新增 Go TUI 工程 `tui/go`，仅作为全屏终端交互外壳。
- TUI 不直接调用 Clash/Mihomo HTTP API，不重新实现订阅、节点、服务、TUN、转换等业务逻辑。
- 所有动作均复用原始 CLI：
  - 订阅：`clashctl sub ...`
  - 代理：`clashctl node ...`
  - 服务/代理环境：`clashctl on` / `clashctl off`
  - TUN：`clashctl tun ...`
  - Web 面板：`clashctl ui`
  - 状态/日志：`clashctl status` / `clashctl log`
- 新增 `clashctl tui` 入口，运行 `$CLASHCTL_HOME/bin/clashctl-tui`。
- 保持 `clashctl ui` / `clashui` 原 Web 面板语义不变。
- 安装流程在存在预编译 Linux TUI 产物时复制到 `$CLASHCTL_HOME/bin/clashctl-tui`。
- 新增 Linux `amd64` / `386` 预编译产物。
