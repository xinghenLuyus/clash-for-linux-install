# 变更记录

## TUI 框架第一版

- 新增 `clashctl tui` / `clashui` 全屏 TUI 入口，基于 `fzf` 搭建主菜单、预览区和基础动作路由。
- 新增 `scripts/tui/` 框架目录：
  - `core.sh`：TUI 主循环、fzf 参数、安装后入口检测。
  - `views.sh`：总览、订阅、节点、日志、Mixin、Web UI、内核服务等预览渲染。
  - `actions.sh`：订阅、节点、TUN、日志、Mixin、Web UI、服务管理等动作路由。
- 原 `clashctl ui` 的 Web 面板地址功能迁移为 `clashctl webui`。
- 保留 `clashctl ui` 兼容入口，并将其导向新的 TUI。
- 安装流程新增 `fzf` 检测：
  - 检测到 `fzf` 且处于交互终端时，可选择安装后直接进入 TUI。
  - 未检测到 `fzf` 时，回退到原 Web UI/CLI 流程。
- 安装复制流程新增 `scripts/tui` 目录同步。
- 新增 `scripts/lib/api.sh`，预留 Clash/Mihomo HTTP API 公共封装，后续用于 Rules、Connections、Traffic、Providers 等 TUI 页面。
- Docker 验证中发现 `fzf` preview 子进程依赖导出的 `CLASHCTL_HOME`，已在 TUI 入口补充 `export CLASHCTL_HOME` 兜底。
- 参考 Clash Verge Rev 的 `calcuProxies` 数据模型扩展 API 层：
  - 将 Clash HTTP API 的 base、secret、URL 编码、GET/POST/PUT/PATCH/DELETE 统一到 `scripts/lib/api.sh`。
  - 新增代理组、普通节点、provider 节点、节点选择、节点测速、组测速、内核升级等公共函数。
  - `node.sh` 改为复用公共 API 层，保留原有交互逻辑。
  - 修复节点名 URL 编码，改为按 UTF-8 字节编码，避免中文节点名生成错误 `%FFFFFFFF...`。
  - 策略组排序对齐 Clash Verge Rev：普通策略组在前，`GLOBAL.all` 引用的策略组移动到后段。
  - 节点展示补充 provider 与 UDP/XUDP/TFO/MPTCP/SMUX 等能力标签。
  - `upgrade.sh` 改为复用 `api_upgrade`，不再手写控制器 API 请求。
- 调整 `clashctl ui` 兼容入口：
  - 进入时明确提示“原 clashctl ui 已更名为 clashctl webui。”
  - 若未安装 `fzf`，直接显示错误并退出，不再回退到 Web UI，避免歧义。
- 调整 `clashui` 直接命令：
  - 直接输入 `clashui` 时不显示旧 `clashctl ui` 改名提示，直接进入 TUI。
  - 仅通过 `clashctl ui` 旧子命令进入时显示兼容提示。
- 收敛 Web UI 入口：
  - 官方入口保留 `clashctl webui`。
  - 底层显示逻辑迁为内部 `_webui_show`。
  - 删除直接 Web UI 用户入口，避免出现两个 Web UI 命令。
- 新增 `scripts/lib/proxy.sh` 作为 Clash Verge 风格代理共享层：
  - 统一策略组列表、组内节点、provider 信息、能力标签、节点切换、组测速、节点测速、排序/过滤和 preview。
  - `clashctl node` 改为走 `proxy_cli_node`，避免 CLI node 与 TUI Proxies 出现两套体验。
- 新增 `scripts/tui/proxies.sh`：
  - TUI Proxies 页面直接复用 `proxy.sh`，不再调用旧 `clashnode` 交互。
  - 支持策略组列表、组 preview、组内节点选择、节点 preview 和切换。
  - 支持 `Ctrl-D` 对当前策略组或当前节点测速，`Ctrl-R` 刷新策略组列表，fzf 原生搜索过滤节点。
- 新增 `scripts/tui/settings.sh`：
  - Settings 页面按 Verge 风格组织。
  - Proxy Environment 复用 `clashon` / `clashoff`，保持与 CLI 行为一致。
  - TUN Mode 复用 `tunon` / `tunoff`。
  - Secret、Mixin、Runtime、Web UI 复用既有命令/函数。
  - Ports / Allow LAN / DNS 先保留入口，后续通过 `mixin.yaml + _merge_config_restart` 实现。
- 新增 Overview live config 与 drift 检测：
  - `api.sh` 新增 `api_configs`，读取 Clash/Mihomo 当前运行配置。
  - 新增 `scripts/lib/drift.sh`，对比 `runtime.yaml` 与 live core 配置。
  - 新增 `scripts/tui/overview.sh`，Overview 优先展示 live API 状态；服务未运行或 API 不可用时回退展示 `runtime.yaml`。
  - 当 live 配置与 `runtime.yaml` 不一致时，仅在 TUI 中提示字段差异，不自动覆盖任何配置。
  - Drift 提示用于解释 Web UI 或外部 API 临时修改运行态后，重启/切订阅/重新合并会回到 `mixin.yaml/runtime.yaml` 的权威配置链路。
- 新增 fzf app shell：
  - 新增 `scripts/tui/frame.sh`，统一页面路由、顶部状态标题、右侧 preview 和动态操作提示。
  - 新增 `scripts/tui/app.sh`，作为常驻式 TUI 主循环。
  - `clashui` / `clashctl tui` 现在进入统一页面导航；页面操作结束后回到 TUI，不再额外要求手动确认返回。
  - 主界面按 Clash Verge-like 信息架构组织为 Overview / Profiles / Proxies / Logs / Settings / Core / Web UI。

## Go TUI 原型

- 新增 `dev-go` 分支上的 Go TUI 原型，目标是替代脆弱的 fzf 拼装层，保留轻量、无第三方 Go 依赖的实现方式。
- 新增 `tui/go/go.mod` 与 `tui/go/cmd/clashctl-tui/main.go`：
  - 使用 Go 标准库直接控制终端 raw mode、alternate screen、键盘输入和全屏渲染。
  - 页面布局固定为顶部状态栏、左侧菜单、右侧内容预览、底部动态操作区。
  - 支持 `q` / `Esc` / `Ctrl-C` 退出，`↑↓` 切换页面，`r` 刷新，`Enter` 执行当前页面动作。
  - 当前阶段页面数据仍复用 `scripts/lib` 与 `scripts/tui/*` 的服务函数，但输出会被捕获进 Go TUI 内容区，避免普通 CLI 输出直接散落在终端。
- 新增 `tui/go/build-tui.sh`，用于构建 `bin/clashctl-tui`。
- `clashctl tui` 与 `clashui` 入口改为优先运行 `bin/clashctl-tui`。
- 安装流程新增可选复制：源码目录存在可执行 `bin/clashctl-tui` 时，会安装到目标 `bin/`。
- 新增 `_tui_status_line`，为 Go TUI 顶部栏提供运行态摘要。
- `.gitignore` 新增 `bin/`，避免把本地构建产物提交进仓库。

## Go TUI 收敛

- Go 工程整体移入 `tui/go/`，避免在项目根目录堆放 TUI 工程文件。
- 删除 fzf TUI 交互适配层，以及 `clashctl node` / proxy 共享层中的 fzf 可选选择器，仅保留 Go TUI、编号式 CLI 和页面展示所需的 shell 服务/渲染函数。
- `clashctl tui` / `clashui` 不再回退到 fzf；如果 `bin/clashctl-tui` 不存在或当前系统不可执行，会明确报错并提示先构建。
- 安装后的 TUI 引导从检测 `fzf` 改为检测可运行的 Go TUI 二进制。
- Go TUI 的 `Enter` 暂时收敛为当前页面刷新，不再跳出 TUI 调用旧 fzf/CLI 交互，后续动作菜单会在 Go 内部逐项实现。
- `tui/go/build-tui.sh` 使用项目内 `.cache/go-build` 作为 Go 构建缓存，避免写入用户目录。
- 已使用本机 Go 构建当前 macOS 开发二进制 `bin/clashctl-tui`，并交叉编译 Linux 测试产物 `bin/clashctl-tui-linux-amd64` 与 `bin/clashctl-tui-linux-386`。
- `tui/go/build-tui.sh` 已移动到 Go TUI 工程目录，并支持 `local`、`linux-amd64`、`linux-386`、`linux`、`all` 构建目标。
- 修复 Ubuntu 终端中 TUI 顶部栏与内容区频闪：
  - 页面内容与状态栏改为缓存，只在启动、切页、手动刷新时重新拉取 shell 服务函数。
  - 渲染过程不再执行 Clash/shell 查询。
  - 普通刷新不再整屏清空，改为回到左上角逐行覆盖并清理行尾。
  - 单次渲染改为先组装 buffer 再一次性写入终端，减少终端重绘撕裂。
