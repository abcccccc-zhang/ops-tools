package cmd

import (
	"deliverops/web" // 注意替换成你的模块路径

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 Web 服务",
	Run: func(cmd *cobra.Command, args []string) {
		web.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
