#!/usr/bin/env bash

clashhelp() {
  cat <<EOF

Usage:
  clashctl COMMAND [OPTIONS]

Commands:
  on                    开启代理
  off                   关闭代理
  status                内核状态
  tui                   全屏 TUI 管理界面
  webui                 Web 面板地址
  ui                    兼容入口，等同于 tui
  sub                   订阅管理
  node                  节点切换
  tun                   Tun 模式
  mixin                 Mixin 配置
  secret                Web 密钥
  log                   查看日志
  upgrade               升级内核

Global Options:
  -h, --help            显示帮助信息

For more help on how to use clashctl, head to https://github.com/nelvko/clash-for-linux-install
EOF
}
