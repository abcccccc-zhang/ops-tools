package cmd

import (
	"deliverops/cloudapi"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	bucketName string
	filePath   string
)
var listCmd = &cobra.Command{
	Use:   "list [bucketName]",
	Short: "列出所有 BOS 存储桶，指定 bucketName 可列对象",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cloudapi.ListBuckets(args...); err != nil {
			fmt.Println("❌ 执行失败:", err)
		}
	},
}

var expireFlag string

var downloadURLCmd = &cobra.Command{
	Use:   "gendownload-url [bucket/object]",
	Short: "生成指定 BOS 对象的下载链接（默认 30 分钟）",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fullPath := strings.TrimPrefix(args[0], "/")
		fullPath = strings.TrimSuffix(fullPath, "/")

		parts := strings.SplitN(fullPath, "/", 2)
		if len(parts) < 2 {
			fmt.Println("❌ 输入格式应为: <bucket>/<objectKey>")
			return
		}

		bucket := parts[0]
		objectKey := parts[1]

		expireSeconds := 1800 // 默认 30 分钟

		if expireFlag != "" {
			if dur, err := time.ParseDuration(expireFlag); err == nil {
				expireSeconds = int(dur.Seconds())
			} else if sec, err := strconv.Atoi(expireFlag); err == nil {
				expireSeconds = sec
			} else {
				fmt.Println("❌ 无效的过期时间格式，支持如 10m、1h、86400")
				return
			}
		}

		url, err := cloudapi.GenerateDownloadURL(bucket, objectKey, expireSeconds)
		if err != nil {
			fmt.Println("❌ 生成下载链接失败:", err)
			return
		}

		fmt.Println("✅ 下载链接:")
		fmt.Println(url)
	},
}

func init() {
	downloadURLCmd.Flags().StringVarP(&expireFlag, "expire", "e", "", "过期时间，如 10m, 1h, 1d 或纯数字秒数")
	rootCmd.AddCommand(downloadURLCmd)
	rootCmd.AddCommand(listCmd)
}
