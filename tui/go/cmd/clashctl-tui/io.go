package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (a *app) capture(fn string) string {
	cmd := exec.Command("bash", "-lc", fmt.Sprintf(". %q/scripts/cmd/clashctl.sh >/dev/null 2>&1; %s 2>&1", a.home, fn))
	cmd.Env = append(os.Environ(), "CLASHCTL_HOME="+a.home)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil && out.Len() == 0 {
		return err.Error()
	}
	return out.String()
}

func (a *app) logsPreview() string {
	logPath := a.logPath()
	if logPath == "" {
		return "最近日志\n\n未找到日志路径。"
	}
	lines, err := readTailLines(logPath, 128*1024, 160)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("最近日志\n\n日志文件：%s\n暂无日志文件。服务启动后会自动生成。", logPath)
		}
		return fmt.Sprintf("最近日志\n\n日志文件：%s\n读取失败：%v", logPath, err)
	}
	if len(lines) == 0 {
		return fmt.Sprintf("最近日志\n\n日志文件：%s\n暂无日志。", logPath)
	}
	return fmt.Sprintf("最近日志\n日志文件：%s\n\n%s", logPath, strings.Join(lines, "\n"))
}

func (a *app) logPath() string {
	kernel := a.kernel()
	resources := filepath.Join(a.home, "resources")
	candidates := []string{
		filepath.Join(resources, kernel+".log"),
		filepath.Join("/var/log", kernel+".log"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return candidates[0]
}

func readTailLines(path string, maxBytes int64, maxLines int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	start := int64(0)
	if size > maxBytes {
		start = size - maxBytes
	}
	if _, err := f.Seek(start, 0); err != nil {
		return nil, err
	}
	buf := make([]byte, size-start)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return nil, err
	}
	text := strings.TrimRight(string(buf[:n]), "\n")
	if text == "" {
		return nil, nil
	}
	lines := strings.Split(text, "\n")
	if start > 0 && len(lines) > 0 {
		lines = lines[1:]
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines, nil
}

func readOptionalTail(path string, maxBytes int64, maxLines int) []string {
	lines, err := readTailLines(path, maxBytes, maxLines)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"暂无日志文件。"}
		}
		return []string{"读取失败：" + err.Error()}
	}
	if len(lines) == 0 {
		return []string{"暂无日志。"}
	}
	return lines
}

func readFilePreview(path string, maxLines int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "文件不存在：" + path
		}
		return "读取失败：" + err.Error()
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}
	return strings.Join(lines, "\n")
}
