package cloudapi

import (
	"fmt"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

// $env:BAIDU_BOS_AK=""
// $env:BAIDU_BOS_SK=""
// $env:BAIDU_BOS_ENDPOINT="https://su.bcebos.com" 如果不用https，生成出来得连接是http得

// ConnectBaiduBOS 初始化 BOS 客户端
func ConnectBaiduBOS() (*bos.Client, error) {
#环境变量还是写死，自己决定
	ak := ""
	sk := ""
	endpoint := "https://su.bcebos.com"

	if ak == "" || sk == "" || endpoint == "" {
		return nil, fmt.Errorf("missing required BOS credentials or endpoint")
	}

	clientConfig := bos.BosClientConfiguration{
		Ak:               ak,
		Sk:               sk,
		Endpoint:         endpoint,
		RedirectDisabled: false,
	}

	client, err := bos.NewClientWithConfig(&clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create BOS client: %w", err)
	}

	return client, nil
}

// ListBuckets 列出当前账户的所有 BOS 存储桶
// 如果提供 bucketName，则列出该 bucket 内的对象
func ListBuckets(bucketPath ...string) error {
	client, err := ConnectBaiduBOS()
	if err != nil {
		return fmt.Errorf("BOS 连接失败: %v", err)
	}

	res, err := client.ListBuckets()
	if err != nil {
		return fmt.Errorf("列出 Buckets 失败: %v", err)
	}

	fmt.Printf("👤 所有者信息：ID=%s, Name=%s\n", res.Owner.Id, res.Owner.DisplayName)

	var filterBucket string
	var prefix string
	if len(bucketPath) > 0 {
		path := strings.TrimSuffix(bucketPath[0], "/")
		parts := strings.SplitN(path, "/", 2)
		filterBucket = parts[0] // Bucket 名称
		if len(parts) > 1 {
			prefix = parts[1] + "/" // 剩下的部分作为 prefix
		}
	}

	count := 0
	for _, b := range res.Buckets {
		if filterBucket != "" && !strings.EqualFold(b.Name, filterBucket) {
			continue
		}
		count++
		fmt.Printf("\n🪣 Bucket #%d\n", count)
		fmt.Println("   名称       :", b.Name)
		fmt.Println("   地区       :", b.Location)
		fmt.Println("   创建时间   :", b.CreationDate)

		if filterBucket != "" {
			if err := listObjectsInBucket(client, b.Name, prefix); err != nil {
				fmt.Println("   ⚠️ 列出对象失败:", err)
			}
		}
	}

	if count == 0 {
		if filterBucket != "" {
			fmt.Printf("⚠️ 没有找到名称为 '%s' 的 Bucket\n", filterBucket)
		} else {
			fmt.Println("⚠️ 当前账户没有任何 Bucket")
		}
	}

	return nil
}

// listObjectsInBucket 列出指定 bucket 下的对象
func listObjectsInBucket(client *bos.Client, bucketName, prefix string) error {
	fmt.Println("   📂 对象列表:")

	args := &api.ListObjectsArgs{
		Prefix:    prefix,
		Delimiter: "/", // 只显示一级
		MaxKeys:   1000,
	}

	res, err := client.ListObjects(bucketName, args)
	if err != nil {
		return err
	}

	// 打印文件对象，显示完整路径
	for i, obj := range res.Contents {
		fmt.Printf("      #%d: %s/%s (大小: %d, 最后修改: %s)\n", i+1, bucketName, obj.Key, obj.Size, obj.LastModified)
	}

	// 打印一级目录，显示完整路径
	for _, dir := range res.CommonPrefixes {
		fmt.Printf("      📁 %s/%s\n", bucketName, dir.Prefix) // dir 本身就是 prefix + 子目录名 + "/"
	}

	if len(res.Contents) == 0 && len(res.CommonPrefixes) == 0 {
		fmt.Println("      空")
	}

	return nil
}

// GenerateDownloadURL 生成指定对象的下载链接
func GenerateDownloadURL(bucketName, objectKey string, expireSeconds int) (string, error) {
	client, err := ConnectBaiduBOS()
	if err != nil {
		return "", fmt.Errorf("BOS 连接失败: %v", err)
	}
	_, err = client.GetObjectMeta(bucketName, objectKey)
	if err != nil {
		return "", fmt.Errorf("检查对象失败: %v", err)
	}
	// #不配置时系统默认值为1800秒。如果要设置为永久不失效的时间，可以将expirationInSeconds参数设置为-1，不可设置为其他负数。
	url := client.BasicGeneratePresignedUrl(bucketName, objectKey, expireSeconds)
	return url, nil
}
