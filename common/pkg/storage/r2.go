package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/my-chat/common/pkg/config"
)

// R2Storage Cloudflare R2 存储客户端
type R2Storage struct {
	client         *s3.Client
	bucketName     string
	exportEndpoint string
}

// NewR2Storage 创建 R2 存储客户端
func NewR2Storage(cfg config.R2Configuration) (*R2Storage, error) {
	// 创建自定义端点解析器
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               cfg.Endpoint,
					SigningRegion:     cfg.Region,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		})

	// 配置 AWS SDK
	awsCfg := aws.Config{
		Region:                      cfg.Region,
		EndpointResolverWithOptions: customResolver,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(awsCfg)

	return &R2Storage{
		client:         client,
		bucketName:     cfg.BucketName,
		exportEndpoint: cfg.ExportEndpoint,
	}, nil
}

// UploadFile 上传文件到 R2
func (r *R2Storage) UploadFile(ctx context.Context, data []byte, filename string) (string, error) {
	// 生成文件路径: YYYY/MM/DD/UUID-filename
	key := fmt.Sprintf("%s/%s-%s", time.Now().Format("2006/01/02"), uuid.New().String(), filename)

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(getContentType(filename)),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to R2: %w", err)
	}
	return key, nil
}

// DownloadFile 从 R2 下载文件
func (r *R2Storage) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from R2: %w", err)
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

// DeleteFile 从 R2 删除文件
func (r *R2Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}
	return nil
}

// ListFiles 列出 R2 中的文件
func (r *R2Storage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	result, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files from R2: %w", err)
	}
	var keys []string
	for _, obj := range result.Contents {
		if obj.Key != nil {
			keys = append(keys, *obj.Key)
		}
	}
	return keys, nil
}

// GetFileURL 获取文件的公开访问 URL
func (r *R2Storage) GetFileURL(key string) string {
	return fmt.Sprintf("https://%s/%s", r.exportEndpoint, key)
}

// getContentType 根据文件扩展名返回 MIME 类型
func getContentType(filename string) string {
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
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/octet-stream"
	}
}
