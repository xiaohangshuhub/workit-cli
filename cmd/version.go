package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "v1.0.10" // 改为实际发布的版本号
	BuildTime = "unknown"
	GitCommit = "unknown"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("workit version %s\n", Version)
			fmt.Printf("  BuildTime: %s\n", BuildTime)
			fmt.Printf("  GitCommit: %s\n", GitCommit)
		},
	}
)
