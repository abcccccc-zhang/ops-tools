package main

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

// Config 配置结构体，包含 cert_id
type Config struct {
	CertID     string `yaml:"cert_id"`
	CertPath   string `yaml:"cert_path"`
	SecretId   string `yaml:"SecretId"`
	SecretKey  string `yaml:"SecretKey"`
	Endpoint   string `yaml:"Endpoint"`
	DomainName string `yaml:"DomainName"`
	ReplaceDay int    `yaml:"ReplaceDay"`
}

// 添加轮询逻辑，直到证书状态为“已签发”
func waitForCertificateIssued(client *ssl.Client, certID string, maxWaitMinutes int, pollIntervalSeconds int) error {
	timeout := time.After(time.Duration(maxWaitMinutes) * time.Minute)
	ticker := time.NewTicker(time.Duration(pollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for certificate to be issued")
		case <-ticker.C:
			// 发起 DescribeCertificate 请求
			request := ssl.NewDescribeCertificateRequest()
			request.CertificateId = common.StringPtr(certID)
			response, err := client.DescribeCertificate(request)
			if err != nil {
				return fmt.Errorf("failed to describe certificate: %v", err)
			}

			// 检查证书状态
			if response.Response != nil {
				statusName := *response.Response.StatusName
				fmt.Printf("Current certificate status: %s\n", statusName)
				if statusName == "已颁发" {
					fmt.Println("Certificate has been issued successfully.")
					return nil
				}
			} else {
				fmt.Println("No response or malformed response from DescribeCertificate.")
			}
		}
	}
}
func main() {
	// 读取配置文件，获取 cert_id
	config, err := loadConfig("./cert_config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 从配置中获取 cert_id
	certID := config.CertID
	if certID == "" {
		log.Fatalf("No cert_id found in config.yaml")
		os.Exit(1)
	}
	cert_path := config.CertPath
	if cert_path == "" {
		log.Fatalf("No cert_path for down cert found in config.yaml")
		os.Exit(1)
	}
	SecretId := config.SecretId
	if SecretId == "" {
		log.Fatalf("No SecretId for down cert found in config.yaml")
		os.Exit(1)
	}
	SecretKey := config.SecretKey
	if SecretKey == "" {
		log.Fatalf("No SecretKey for down cert found in config.yaml")
		os.Exit(1)
	}
	Endpoint := config.Endpoint
	if Endpoint == "" {
		log.Fatalf("No Endpoint for down cert found in config.yaml")
		os.Exit(1)
	}
	ReplaceDay := config.ReplaceDay
	if ReplaceDay == 0 {
		log.Fatalf("No ReplaceDay for down cert found in config.yaml")
		os.Exit(1)
	}
	// 创建认证对象
	credential := common.NewCredential(
		SecretId,  // 填写你的 SecretId
		SecretKey, // 填写你的 SecretKey
	)

	// 创建客户端配置
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = Endpoint

	// 实例化SSL客户端
	client, _ := ssl.NewClient(credential, "", cpf)

	// 获取证书详情
	request := ssl.NewDescribeCertificateRequest()
	request.CertificateId = common.StringPtr(certID)

	// 发起 DescribeCertificate 请求
	response, err := client.DescribeCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s\n", err)
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	// 输出回包（JSON格式）
	fmt.Printf("DescribeCertificate response: %s\n", response.ToJsonString())
	// 提取证书过期日期（CertEndTime）
	if response.Response != nil {
		StatusName := *response.Response.StatusName
		fmt.Printf("The certificate Status is: %s\n", StatusName)
		if StatusName == "已过期" || StatusName == "已吊销" {
			// 如果证书已过期，更新证书并写入新的 certId
			fmt.Println("证书已过期，开始更新证书")
			newCertID := NewCertificate(client)

			// 更新 config.yaml 中的 cert_id 字段
			err = updateConfig("./cert_config.yaml", newCertID)
			if err != nil {
				log.Fatalf("Failed to update config.yaml: %v", err)
			}
			// 输出更新后的证书 ID
			fmt.Printf("配置已更新，新证书 ID: %s\n", newCertID)
			maxWaitMinutes := 30      // 最大等待时间为 30 分钟
			pollIntervalSeconds := 10 // 每 10 秒检查一次
			fmt.Println("Waiting for certificate to be issued...")
			err := waitForCertificateIssued(client, newCertID, maxWaitMinutes, pollIntervalSeconds)
			if err != nil {
				log.Fatalf("Failed to wait for certificate issuance: %v", err)
			}

			// 下载证书
			downCert(client, cert_path, newCertID)
		} else {
			// 如果证书未过期
			fmt.Println("证书未过期")
		}
		// 获取证书的过期时间并进行比较
		certEndTime := *response.Response.CertEndTime
		expiryTime, err := time.Parse("2006-01-02 15:04:05", certEndTime)
		if err != nil {
			log.Fatalf("Failed to parse CertEndTime: %v", err)
		}
		// 计算当前时间与证书过期时间的差值
		timeUntilExpiry := time.Until(expiryTime)
		fmt.Printf("证书将于 %s 过期, 距离现在还剩 %s\n", certEndTime, timeUntilExpiry)
		// 如果证书在30天内过期，触发更新操作
		if timeUntilExpiry < time.Duration(ReplaceDay)*24*time.Hour {
			fmt.Printf("证书将在 %d 天内过期，准备更新证书...", ReplaceDay)
			newCertID := NewCertificate(client)
			// 更新配置文件
			err := updateConfig("./cert_config.yaml", newCertID)
			if err != nil {
				log.Fatalf("Failed to update config.yaml: %v", err)
			}
			maxWaitMinutes := 30      // 最大等待时间为 30 分钟
			pollIntervalSeconds := 10 // 每 10 秒检查一次
			fmt.Println("Waiting for certificate to be issued...")
			err = waitForCertificateIssued(client, newCertID, maxWaitMinutes, pollIntervalSeconds)
			if err != nil {
				log.Fatalf("Failed to wait for certificate issuance: %v", err)
			}

			// 下载证书
			downCert(client, cert_path, newCertID)
		} else {
			fmt.Printf("证书在未来 %d 天内不会过期", ReplaceDay)
		}
	} else {
		fmt.Println("Error: Response is empty or malformed")
	}
}

// 将证书 ID 写入 config.yaml 中，并更新 cert_id 字段
func updateConfig(configPath string, newCertID string) error {
	// 读取现有配置
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// 更新 cert_id 字段
	config.CertID = newCertID

	// 写入更新后的配置到文件
	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, configData, 0644) // 使用 os.WriteFile 替代 ioutil.WriteFile
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// 申请新的证书并返回证书 ID
func NewCertificate(client *ssl.Client) string {
	// 实例化一个请求对象
	request := ssl.NewApplyCertificateRequest()
	config, err := loadConfig("./cert_config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	DomainName := config.DomainName
	if DomainName == "" {
		log.Fatalf("No DomainName for down cert found in config.yaml")
		os.Exit(1)
	}
	// 设置更新证书的相关参数
	request.DvAuthMethod = common.StringPtr("DNS_AUTO")
	request.DomainName = common.StringPtr(DomainName)

	// 发起 ApplyCertificate 请求
	response, err := client.ApplyCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s\n", err)
		return ""
	}
	if err != nil {
		log.Fatalf("Error applying certificate: %v", err)
	}

	// 提取新的证书 ID
	certID := *response.Response.CertificateId
	fmt.Printf("新的证书 ID: %s\n", certID)
	return certID
}

// 读取YAML配置文件
func loadConfig(configPath string) (Config, error) {
	var config Config
	// 打开配置文件
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	// 读取配置文件内容
	content, err := io.ReadAll(file) // 使用 io.ReadAll 替代 ioutil.ReadAll
	if err != nil {
		return config, err
	}

	// 解析 YAML 配置文件
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// 下载证书
func downCert(client *ssl.Client, certPath string, CertID string) {

	request := ssl.NewDownloadCertificateRequest()
	request.CertificateId = common.StringPtr(CertID)
	// 返回的resp是一个DownloadCertificateResponse的实例，与请求对象对应
	response, err := client.DownloadCertificate(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	// fmt.Printf("%s", response.ToJsonString())
	// 获取返回的证书内容和类型
	content := response.Response.Content
	if content == nil {
		log.Println("No certificate content returned.")
		return
	}

	// 解码 Base64 编码的证书内容
	decodedContent, err := base64.StdEncoding.DecodeString(*content)
	if err != nil {
		log.Fatalf("Failed to decode base64 content: %v", err)
	}

	// 创建目录（如果不存在）
	err = os.MkdirAll(certPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// 将解码后的证书内容保存为 ZIP 文件
	err = os.WriteFile(certPath+"/certificate.zip", decodedContent, 0644)
	if err != nil {
		log.Fatalf("Failed to write certificate to file: %v", err)
	}

	fmt.Printf("Certificate saved to: %s/certificate.zip\n", certPath)

	// 解压 ZIP 文件到指定路径
	err = unzip(certPath+"/certificate.zip", certPath)
	if err != nil {
		log.Fatalf("Failed to unzip certificate: %v", err)
	}

	// 输出解压后的文件路径
	fmt.Printf("Certificate has been extracted to: %s\n", certPath)
}

func unzip(zipFile, destDir string) error {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	// 遍历 ZIP 文件中的所有文件
	for _, file := range r.File {
		// 获取文件的完整路径
		filePath := filepath.Join(destDir, file.Name)

		// 如果是目录，则创建目录
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}

		// 否则是文件，打开并解压
		err := extractFile(file, filePath)
		if err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
	}

	return nil
}

// extractFile 解压单个文件
func extractFile(file *zip.File, destPath string) error {
	// 打开源文件
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file inside zip: %v", err)
	}
	defer rc.Close()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer destFile.Close()

	// 将文件内容复制到目标文件
	_, err = io.Copy(destFile, rc)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}
