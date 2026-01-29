package handler

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-chat/common/pkg/storage"
	"github.com/redis/go-redis/v9"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	r2        *storage.R2Storage
	redis     *redis.Client
	rateLimit int // 每小时每用户最大上传次数，0 表示不限制
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(r2 *storage.R2Storage, redisClient *redis.Client, rateLimit int) *UploadHandler {
	return &UploadHandler{
		r2:        r2,
		redis:     redisClient,
		rateLimit: rateLimit,
	}
}

// UploadRequest 上传请求
type UploadRequest struct {
	Filename string `json:"filename"` // 文件名
	Data     string `json:"data"`     // Base64 编码的文件数据
	MimeType string `json:"mime_type,omitempty"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	Fid    string `json:"fid"`    // 文件ID
	Name   string `json:"name"`   // 文件名
	Size   int64  `json:"size"`   // 文件大小
	Mime   string `json:"mime"`   // MIME 类型
	SHA256 string `json:"sha256"` // 文件哈希
	URL    string `json:"url"`    // 访问 URL
	Key    string `json:"key"`    // 存储 Key
}

// MaxFileSize 最大文件大小 (20MB)
const MaxFileSize = 20 * 1024 * 1024

// checkRateLimit 检查上传频率限制
// 返回: 是否允许上传, 剩余次数, 错误
func (h *UploadHandler) checkRateLimit(ctx context.Context, uid string) (bool, int, error) {
	if h.rateLimit <= 0 || h.redis == nil {
		return true, -1, nil // 不限制
	}

	key := fmt.Sprintf("upload_rate:%s", uid)

	// 获取当前计数
	count, err := h.redis.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, 0, err
	}

	if count >= h.rateLimit {
		return false, 0, nil
	}

	// 增加计数
	pipe := h.redis.Pipeline()
	pipe.Incr(ctx, key)
	// 如果是新 key，设置 1 小时过期
	pipe.Expire(ctx, key, time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	return true, h.rateLimit - count - 1, nil
}

// Upload 处理文件上传
func (h *UploadHandler) Upload(c *gin.Context) {
	if h.r2 == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage service unavailable"})
		return
	}

	// 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 检查上传频率限制
	allowed, remaining, err := h.checkRateLimit(c.Request.Context(), uid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "upload rate limit exceeded",
			"message": fmt.Sprintf("maximum %d uploads per hour", h.rateLimit),
		})
		return
	}

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 验证文件名
	if req.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	// Base64 解码
	fileData, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 data"})
		return
	}

	// 检查文件大小
	if len(fileData) > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 20MB)"})
		return
	}

	// 计算 SHA256
	hash := sha256.Sum256(fileData)
	sha256Hex := hex.EncodeToString(hash[:])

	// 生成文件ID
	fid := uuid.New().String()

	// 生成新文件名
	ext := filepath.Ext(req.Filename)
	newFilename := fid + ext

	// 上传到 R2
	key, err := h.r2.UploadFile(c.Request.Context(), fileData, newFilename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	// 获取 MIME 类型
	mime := req.MimeType
	if mime == "" {
		mime = getMimeType(req.Filename)
	}

	resp := UploadResponse{
		Fid:    fid,
		Name:   req.Filename,
		Size:   int64(len(fileData)),
		Mime:   mime,
		SHA256: sha256Hex,
		URL:    h.r2.GetFileURL(key),
		Key:    key,
	}

	// 添加剩余次数到响应头
	if remaining >= 0 {
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", h.rateLimit))
	}

	c.JSON(http.StatusOK, resp)
}

// UploadMultipart 处理 multipart 文件上传
func (h *UploadHandler) UploadMultipart(c *gin.Context) {
	if h.r2 == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage service unavailable"})
		return
	}

	// 获取用户ID
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 检查上传频率限制
	allowed, remaining, err := h.checkRateLimit(c.Request.Context(), uid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "upload rate limit exceeded",
			"message": fmt.Sprintf("maximum %d uploads per hour", h.rateLimit),
		})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 20MB)"})
		return
	}

	// 读取文件内容
	fileData := make([]byte, header.Size)
	if _, err := file.Read(fileData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// 计算 SHA256
	hash := sha256.Sum256(fileData)
	sha256Hex := hex.EncodeToString(hash[:])

	// 生成文件ID
	fid := uuid.New().String()

	// 生成新文件名
	ext := filepath.Ext(header.Filename)
	newFilename := fid + ext

	// 上传到 R2
	key, err := h.r2.UploadFile(c.Request.Context(), fileData, newFilename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	resp := UploadResponse{
		Fid:    fid,
		Name:   header.Filename,
		Size:   header.Size,
		Mime:   getMimeType(header.Filename),
		SHA256: sha256Hex,
		URL:    h.r2.GetFileURL(key),
		Key:    key,
	}

	// 添加剩余次数到响应头
	if remaining >= 0 {
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", h.rateLimit))
	}

	c.JSON(http.StatusOK, resp)
}

// getMimeType 根据文件扩展名获取 MIME 类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".zip":
		return "application/zip"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
