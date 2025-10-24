package web

import (
	"deliverops/cloudapi"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
)

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

type PackageItem struct {
	Section string `json:"section"`
	Path    string `json:"path"`
}
type GenerateDownloadURLRequest struct {
	Path   string `json:"path" binding:"required"`
	Expire string `json:"expire"`
	// BucketName string `json:"bucketname"`
}
type GenerateDownloadURLResponse struct {
	URL string `json:"url"`
}

func HandleGenerateDownloadURL(c *gin.Context) {
	var req GenerateDownloadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 解析过期时间
	seconds, err := parseExpireToSeconds(req.Expire)
	if err != nil {
		c.JSON(400, gin.H{"error": "过期时间格式错误: " + err.Error()})
		return
	}

	// 根据 path 和 arch 解析 bucketName 和 objectKey（根据你的业务逻辑）
	bucketName := "premium-deploy"
	objectKey := strings.TrimPrefix(req.Path, "premium-deploy/")

	url, err := cloudapi.GenerateDownloadURL(bucketName, objectKey, seconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Error: fmt.Sprintf("生成失败: %v", err)})
		return
	}
	fmt.Println("下载链接: %s 超时时间：%s", url, seconds)
	c.JSON(200, GenerateDownloadURLResponse{URL: url})
}

// 主入口：处理 POST 请求
func HandleListPackage(c *gin.Context) {
	var req struct {
		Version string `json:"version"` // 分支名
		Type    string `json:"type"`    // single / k8s
		Arch    string `json:"arch"`    // amd64 / arm64
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析失败"})
		return
	}

	fmt.Printf("handleListPackage: version=%s, type=%s, arch=%s\n", req.Version, req.Type, req.Arch)

	// === 单机版 ===
	if req.Type == "single" {
		path := fmt.Sprintf("premium-deploy/01-安装部署/premium-deploy/x86/premium-deploy-%s.tar.gz", req.Version)
		if req.Arch == "arm64" {
			path = fmt.Sprintf("premium-deploy/01-安装部署/premium-deploy/ARM/premium-deploy-%s.tar.gz", req.Version)
		}
		fmt.Println("返回单机版路径:", path)
		c.JSON(http.StatusOK, gin.H{
			"data": []PackageItem{
				{Path: path},
			},
		})
		return
	}

	// === K8S 版 ===
	if req.Type == "k8s" {
		repoURL := "https://gitee.com/oschina/gitee-helm-chart.git"
		branch := req.Version
		localPath := "./git-repo"

		authUser := "zjlll"
		authPass := "Zhang.123"

		fmt.Println("开始处理 K8S 安装包...")

		if err := cloneGitRepo(repoURL, branch, localPath, authUser, authPass); err != nil {
			fmt.Println("Git 拉取失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Git 拉取失败: " + err.Error()})
			return
		}

		configFile := localPath + "/VERSION.toml"
		fmt.Println("解析配置文件:", configFile)

		packages, err := getK8SPackageList(configFile, req.Arch)
		if err != nil {
			fmt.Println("配置文件解析失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "配置文件解析失败: " + err.Error()})
			return
		}

		fmt.Println("最终获取到包列表:")
		for _, pkg := range packages {
			fmt.Println(" -", pkg.Path)
		}

		c.JSON(http.StatusOK, gin.H{"data": packages})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "未知类型"})
}

// Git 拉取或切换分支
func cloneGitRepo(repoURL, version, localPath, authUser, authPass string) error {
	var authMethod *gitHttp.BasicAuth
	if authUser != "" {
		authMethod = &gitHttp.BasicAuth{
			Username: authUser,
			Password: authPass,
		}
	}

	// 如果目录不存在，先完整克隆默认分支
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		fmt.Println("目录不存在，开始克隆默认分支:", repoURL)
		_, err := git.PlainClone(localPath, false, &git.CloneOptions{
			URL:   repoURL,
			Depth: 1,
			Auth:  authMethod,
		})
		if err != nil {
			return fmt.Errorf("克隆仓库失败: %w", err)
		}
	}

	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return fmt.Errorf("打开本地仓库失败: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作树失败: %w", err)
	}

	// 尝试切换到指定分支
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(version),
		Force:  true,
	})
	if err != nil {
		// 分支不存在，尝试 tag
		fmt.Println("分支不存在，尝试 tag:", version)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(version),
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("分支和 tag 都找不到: %w", err)
		}
	}

	// 如果切换的是分支，则拉取最新
	headRef, err := repo.Head()
	if err == nil && headRef.Name().IsBranch() {
		err = w.Pull(&git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: headRef.Name(),
			SingleBranch:  true,
			Force:         true,
			Auth:          authMethod,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			fmt.Println("拉取最新分支失败:", err)
		}
	}

	fmt.Println("Git 同步完成:", version)
	return nil
}

// 解析 toml 获取架构相关 bos 地址
func getK8SPackageList(configFile string, arch string) ([]PackageItem, error) {
	v := viper.New()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var result []PackageItem
	allSections := []string{"poc", "kubeasz", "gitee", "scan", "go", "wiki", "middleware"}

	flattened := v.AllSettings() // map[string]interface{}

	for _, section := range allSections {
		if secData, ok := flattened[section]; ok {
			if secMap, ok := secData.(map[string]interface{}); ok {
				key := fmt.Sprintf("%s_bos", arch)
				if val, ok := secMap[key]; ok {
					if str, ok := val.(string); ok && str != "" {
						str = strings.TrimSpace(str)
						if strings.HasPrefix(str, "bos:") {
							str = strings.TrimPrefix(str, "bos:/")
						}
						result = append(result, PackageItem{
							Section: section,
							Path:    str,
						})
					}
				}
			}
		}
	}

	return result, nil
}

func parseExpireToSeconds(expireStr string) (int, error) {
	if expireStr == "" {
		return 1800, nil // 默认 1800 秒
	}

	unit := strings.ToLower(expireStr[len(expireStr)-1:])
	valueStr := expireStr[:len(expireStr)-1]
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("无效的过期时间: %v", err)
	}

	switch unit {
	case "s":
		return value, nil
	case "m":
		return value * 60, nil
	case "h":
		return value * 3600, nil
	default:
		return 0, fmt.Errorf("不支持的单位: %s", unit)
	}
}
