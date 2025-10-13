package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gopkg.in/ini.v1"
)

// 配置结构体
type Config struct {
	Server struct {
		Port string `ini:"port"`
	}
	XHSDownloader struct {
		URL         string `ini:"url"`
		DownloadDir string `ini:"download_dir"`
	}
	AList struct {
		URL        string `ini:"url"`
		Token      string `ini:"token"`
		UploadPath string `ini:"upload_path"`
		AsTask     string `ini:"as_task"`
	}
	CORS struct {
		AllowOrigin  string `ini:"allow_origin"`
		AllowMethods string `ini:"allow_methods"`
		AllowHeaders string `ini:"allow_headers"`
	}
}

// 请求体结构
type RequestBody struct {
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

// XHS下载器响应结构
type XHSDownloadResponse struct {
	Message string          `json:"message"`
	Data    XHSDownloadData `json:"data"`
}

type XHSDownloadData struct {
	Author        string   `json:"作者昵称"`
	Title         string   `json:"作品标题"`
	DownloadLinks []string `json:"下载地址"`
	ItemType      string   `json:"作品类型"`
	ItemID        string   `json:"作品ID"`
}

// 资源信息结构
type ResourceInfo struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
}

// 上传结果结构
type UploadResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
}

// 最终返回结构
type Response struct {
	Message     string         `json:"message"`
	Download    bool           `json:"download"`
	Upload      bool           `json:"upload"`
	Resources   []ResourceInfo `json:"resources"`
	UploadTasks []UploadResult `json:"upload_tasks"`
}

// ALIST上传响应结构
type AlistUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	} `json:"data"`
}

var config Config

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建Gin路由
	r := gin.Default()

	// 提供静态文件服务
	r.StaticFile("/", "./index.html")

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", config.CORS.AllowOrigin)
		c.Header("Access-Control-Allow-Methods", config.CORS.AllowMethods)
		c.Header("Access-Control-Allow-Headers", config.CORS.AllowHeaders)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 定义POST接口
	r.POST("/api/process", func(c *gin.Context) {
		var req RequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "无效的请求体格式",
			})
			return
		}

		// 提取URL
		urlRegex := regexp.MustCompile(`https?://[^\s"']+`)
		targetURL := urlRegex.FindString(req.Message.Text)
		if targetURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "未找到以http或https开头的URL",
			})
			return
		}

		// 调用XHS下载器
		downloadResponse, err := downloadFromXHS(targetURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Message:  fmt.Sprintf("下载失败: %v", err),
				Download: false,
				Upload:   false,
			})
			return
		}

		// 资源文件
		resources := generateResourcesInfo(config.XHSDownloader.DownloadDir, downloadResponse)

		// 上传到AList
		uploadResults, uploadSuccess := uploadToAlist(resources)

		// 清除资源
		clearDownloadDir(resources)

		// 返回结果
		c.JSON(http.StatusOK, Response{
			Message:     "处理完成",
			Download:    true,
			Upload:      uploadSuccess,
			Resources:   resources,
			UploadTasks: uploadResults,
		})
	})

	// 启动服务器
	fmt.Printf("服务器启动在 http://localhost:%s\n", config.Server.Port)
	if err := r.Run(":" + config.Server.Port); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

// 加载配置
func loadConfig() error {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 直接读取配置项，而不是使用MapTo
	serverSection := cfg.Section("server")
	config.Server.Port = serverSection.Key("port").String()

	xhsSection := cfg.Section("xhs_downloader")
	config.XHSDownloader.URL = xhsSection.Key("url").String()
	config.XHSDownloader.DownloadDir = xhsSection.Key("download_dir").String()

	alistSection := cfg.Section("alist")
	config.AList.URL = alistSection.Key("url").String()
	config.AList.Token = alistSection.Key("token").String()
	config.AList.UploadPath = alistSection.Key("upload_path").String()
	config.AList.AsTask = alistSection.Key("as_task").String()

	corsSection := cfg.Section("cors")
	config.CORS.AllowOrigin = corsSection.Key("allow_origin").String()
	config.CORS.AllowMethods = corsSection.Key("allow_methods").String()
	config.CORS.AllowHeaders = corsSection.Key("allow_headers").String()

	// 设置默认值
	if config.Server.Port == "" {
		config.Server.Port = "9092"
	}
	if config.XHSDownloader.URL == "" {
		config.XHSDownloader.URL = "http://xxx/xhs/detail"
	}
	if config.XHSDownloader.DownloadDir == "" {
		config.XHSDownloader.DownloadDir = "/root/res_saver/xhs_downloader_volume/Download"
	}
	if config.AList.URL == "" {
		config.AList.URL = "http://xxx/api/fs/put"
	}
	if config.CORS.AllowOrigin == "" {
		config.CORS.AllowOrigin = "*"
	}
	if config.CORS.AllowMethods == "" {
		config.CORS.AllowMethods = "POST, OPTIONS"
	}
	if config.CORS.AllowHeaders == "" {
		config.CORS.AllowHeaders = "Origin, Content-Type, Accept"
	}
	if config.AList.AsTask == "" {
		config.AList.AsTask = "true"
	}

	return nil
}

func clearDownloadDir(resources []ResourceInfo) {
	for _, res := range resources {
		if err := os.Remove(res.FilePath); err != nil {
			fmt.Printf("删除文件 %s 失败: %v\n", res.FilePath, err)
		}
	}
}

// 调用XHS下载器
func downloadFromXHS(targetURL string) (*XHSDownloadResponse, error) {
	client := resty.New()

	// 准备请求体
	reqBody := map[string]interface{}{
		"url":      targetURL,
		"download": true,
		"skip":     false,
	}

	// 发送POST请求
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		SetResult(&XHSDownloadResponse{}).
		Post(config.XHSDownloader.URL)

	if err != nil {
		return nil, fmt.Errorf("调用XHS下载器失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("XHS下载器返回非200状态码: %d", resp.StatusCode())
	}

	result, ok := resp.Result().(*XHSDownloadResponse)
	if !ok {
		return nil, fmt.Errorf("解析XHS下载器响应失败")
	}

	return result, nil
}

// 资源文件
func generateResourcesInfo(dir string, response *XHSDownloadResponse) []ResourceInfo {
	var resources []ResourceInfo

	// 获取作者昵称和作品标题
	authorName := response.Data.Author
	itemID := response.Data.ItemID

	var ext string
	if response.Data.ItemType == "图文" {
		ext = ".jpeg"
	} else if response.Data.ItemType == "视频" {
		ext = ".mov"
	} else {
		ext = ".jpeg" // 默认扩展名
	}

	// 清理标题中的特殊字符
	// 处理所有下载地址
	for i, url := range response.Data.DownloadLinks {
		if url == "" {
			continue
		}
		// 生成文件名
		fileName := fmt.Sprintf("%s_%s_%d%s", authorName, itemID, i+1, ext)
		filePath := filepath.Join(dir, fileName)

		// 添加到资源列表
		resources = append(resources, ResourceInfo{
			FileName: fileName,
			FilePath: filePath,
		})
	}

	return resources
}

// 上传到AList
func uploadToAlist(resources []ResourceInfo) ([]UploadResult, bool) {
	var results []UploadResult
	allSuccess := true
	client := resty.New()

	for _, resource := range resources {
		// 读取文件内容
		fileData, err := os.ReadFile(resource.FilePath)
		if err != nil {
			results = append(results, UploadResult{
				Success: false,
				Error:   fmt.Sprintf("读取文件失败: %v", err),
			})
			allSuccess = false
			continue
		}

		// 准备文件路径
		targetPath := fmt.Sprintf("%s/%s", config.AList.UploadPath, url.QueryEscape(resource.FileName))

		// 发送上传请求
		resp, err := client.R().
			SetHeader("Authorization", config.AList.Token).
			SetHeader("File-Path", targetPath).
			SetHeader("As-Task", config.AList.AsTask).
			SetHeader("Content-Type", "application/octet-stream").
			SetHeader("Content-Length", fmt.Sprintf("%d", len(fileData))).
			SetBody(fileData).
			SetResult(&AlistUploadResponse{}).
			Put(config.AList.URL)

		if err != nil {
			results = append(results, UploadResult{
				Success: false,
				Error:   fmt.Sprintf("上传请求失败: %v", err),
			})
			allSuccess = false
			continue
		}

		// 检查响应
		if resp.StatusCode() != http.StatusOK {
			results = append(results, UploadResult{
				Success: false,
				Error:   fmt.Sprintf("上传失败，状态码: %d", resp.StatusCode()),
			})
			allSuccess = false
			continue
		}

		// 解析响应
		var uploadResp AlistUploadResponse
		if err := json.Unmarshal(resp.Body(), &uploadResp); err != nil {
			results = append(results, UploadResult{
				Success: false,
				Error:   fmt.Sprintf("解析上传响应失败: %v", err),
			})
			allSuccess = false
			continue
		}

		if uploadResp.Code != 200 {
			results = append(results, UploadResult{
				Success: false,
				Error:   fmt.Sprintf("上传失败: %s", uploadResp.Message),
			})
			allSuccess = false
			continue
		}

		// 上传成功
		results = append(results, UploadResult{
			Success: true,
			TaskID:  uploadResp.Data.Task.ID,
		})

	}

	return results, allSuccess
}
