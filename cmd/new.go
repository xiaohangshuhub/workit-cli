package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	optTemplate string
	optBranch   string
	optForce    bool

	newCmd = &cobra.Command{
		Use:     "new [name]",
		Short:   "创建新项目",
		Example: "workit new myapp --template https://github.com/xiaohangshuhub/go-workit",
		Args:    cobra.ExactArgs(1),
		RunE:    runNew,
	}
)

func init() {
	newCmd.Flags().StringVarP(&optTemplate, "template", "t", "git@github.com:xiaohangshuhub/go-workit.git", "模板仓库地址")
	newCmd.Flags().StringVarP(&optBranch, "branch", "b", "cli-template", "模板仓库分支")
	newCmd.Flags().BoolVarP(&optForce, "force", "f", false, "强制创建(覆盖已存在目录)")
}

func runNew(cmd *cobra.Command, args []string) error {
	name := args[0]

	// 检查目录
	if !optForce {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			return fmt.Errorf("目录已存在: %s (使用 --force 强制创建)", name)
		}
	}

	fmt.Printf("创建项目: %s\n", name)
	fmt.Printf("模板: %s (%s)\n", optTemplate, optBranch)

	// 克隆模板
	if err := cloneTemplate(name); err != nil {
		return err
	}

	// 初始化项目
	if err := initProject(name); err != nil {
		return err
	}

	showSuccess(name)
	return nil
}

func cloneTemplate(name string) error {
	if optForce {
		os.RemoveAll(name)
	}

	// 检查 git 命令是否可用
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git 命令未找到，请先安装 git")
	}

	fmt.Printf("➤ 开始克隆模板...\n  源: %s\n  分支: %s\n", optTemplate, optBranch)

	// 创建带超时的 context (增加到2分钟)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 使用 --depth 1 和 --progress 参数优化克隆过程
	cmd := exec.CommandContext(ctx, "git", "clone",
		"--depth", "1",
		"--progress",
		"-b", optBranch,
		optTemplate,
		name,
	)

	// 设置环境变量，强制显示进度
	cmd.Env = append(os.Environ(), "GIT_PROGRESS=true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("克隆超时(2分钟)，请检查:\n1. 是否可以访问 %s\n2. 网络连接是否正常", optTemplate)
		}
		// 添加调试信息
		fmt.Printf("Debug - 克隆命令: git clone --depth 1 --progress -b %s %s %s\n", optBranch, optTemplate, name)
		return fmt.Errorf("克隆失败: %v\n建议手动尝试克隆命令检查具体原因", err)
	}

	return nil
}

func initProject(name string) error {
	// 删除.git
	if err := os.RemoveAll(filepath.Join(name, ".git")); err != nil {
		return fmt.Errorf("清理git失败: %v", err)
	}

	// 初始化go.mod
	fmt.Println("➤ 初始化go.mod...")
	cmd := exec.Command("go", "mod", "edit", "-module", name)
	cmd.Dir = name
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("初始化go.mod失败: %v", err)
	}

	//  替换 cmd/ 目录下的 import
	oldImport := optBranch

	if !strings.Contains(optBranch, "cli") {
		oldImport = "github.com/xiaohangshuhub/go-workit"
	}

	cmdPath := filepath.Join(name, "")
	newImport := fmt.Sprintf("%s", name)
	err := replaceCmdImports(cmdPath, oldImport, newImport)
	if err != nil {
		return fmt.Errorf("替换 import 失败: %v", err)
	}
	// 安装依赖
	fmt.Println("➤ 安装依赖...")
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func showSuccess(name string) {
	fmt.Printf("\n✓ 项目创建成功: %s\n", name)
	fmt.Printf("\n运行项目:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  go run cmd/service1/main.go\n")
	fmt.Printf("  浏览器访问:http://localhost:8080/hello\n")
	fmt.Printf("  浏览器访问:http://localhost:8080/swagger/index.html\n")
}
func replaceCmdImports(cmdDir, oldImport, newImport string) error {
	return filepath.WalkDir(cmdDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(content), oldImport) {
				newContent := strings.ReplaceAll(string(content), oldImport, newImport)
				err = os.WriteFile(path, []byte(newContent), 0644)
				if err != nil {
					return err
				}
				fmt.Printf("➤ 替换 import 成功: %s\n", path)
			}
		}
		return nil
	})
}
