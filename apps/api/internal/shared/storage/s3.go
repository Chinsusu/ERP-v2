package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrObjectStorageUnavailable = errors.New("object storage is unavailable")

type S3Config struct {
	Endpoint     string
	Region       string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	UsePathStyle bool
	HTTPClient   *http.Client
}

type S3CompatibleObjectStore struct {
	config S3Config
	client *http.Client
	clock  func() time.Time
}

func NewS3CompatibleObjectStore(config S3Config) *S3CompatibleObjectStore {
	if strings.TrimSpace(config.Region) == "" {
		config.Region = "us-east-1"
	}
	client := config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	return &S3CompatibleObjectStore{
		config: config,
		client: client,
		clock:  func() time.Time { return time.Now().UTC() },
	}
}

func (store *S3CompatibleObjectStore) PutObject(
	ctx context.Context,
	bucket string,
	key string,
	contentType string,
	size int64,
	body io.Reader,
) error {
	if store == nil || body == nil || size <= 0 {
		return ErrObjectStorageUnavailable
	}
	endpoint := strings.TrimSpace(store.config.Endpoint)
	accessKey := strings.TrimSpace(store.config.AccessKey)
	secretKey := strings.TrimSpace(store.config.SecretKey)
	bucket = strings.TrimSpace(bucket)
	key = strings.Trim(strings.TrimSpace(key), "/")
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" || key == "" {
		return ErrObjectStorageUnavailable
	}

	data, err := io.ReadAll(io.LimitReader(body, size+1))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrObjectStorageUnavailable, err)
	}
	if int64(len(data)) != size {
		return fmt.Errorf("%w: object size mismatch", ErrObjectStorageUnavailable)
	}

	objectURL, err := store.objectURL(endpoint, bucket, key)
	if err != nil {
		return err
	}
	payloadHash := sha256Hex(data)
	now := store.clock().UTC()
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL.String(), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrObjectStorageUnavailable, err)
	}
	request.ContentLength = size
	if strings.TrimSpace(contentType) != "" {
		request.Header.Set("Content-Type", strings.TrimSpace(contentType))
	}
	request.Header.Set("X-Amz-Content-Sha256", payloadHash)
	request.Header.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	request.Header.Set("Authorization", store.authorizationHeader(request, payloadHash, now))

	response, err := store.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrObjectStorageUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("%w: put object status %d %s", ErrObjectStorageUnavailable, response.StatusCode, strings.TrimSpace(string(message)))
	}

	return nil
}

func (store *S3CompatibleObjectStore) objectURL(endpoint string, bucket string, key string) (*url.URL, error) {
	if !strings.Contains(endpoint, "://") {
		if store.config.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}
	parsed, err := url.Parse(strings.TrimRight(endpoint, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: invalid endpoint", ErrObjectStorageUnavailable)
	}
	if store.config.UsePathStyle {
		parsed.Path = "/" + escapeS3Path(bucket) + "/" + escapeS3Path(key)
		return parsed, nil
	}

	parsed.Host = bucket + "." + parsed.Host
	parsed.Path = "/" + escapeS3Path(key)
	return parsed, nil
}

func (store *S3CompatibleObjectStore) authorizationHeader(request *http.Request, payloadHash string, now time.Time) string {
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	region := strings.TrimSpace(store.config.Region)
	host := request.URL.Host
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := "host:" + host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"
	canonicalRequest := strings.Join([]string{
		request.Method,
		request.URL.EscapedPath(),
		request.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	credentialScope := dateStamp + "/" + region + "/s3/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(store.config.SecretKey, dateStamp, region), []byte(stringToSign)))

	return "AWS4-HMAC-SHA256 Credential=" + strings.TrimSpace(store.config.AccessKey) + "/" + credentialScope +
		", SignedHeaders=" + signedHeaders + ", Signature=" + signature
}

func escapeS3Path(value string) string {
	parts := strings.Split(value, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func signingKey(secret string, dateStamp string, region string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte("s3"))
	return hmacSHA256(serviceKey, []byte("aws4_request"))
}

func hmacSHA256(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}
