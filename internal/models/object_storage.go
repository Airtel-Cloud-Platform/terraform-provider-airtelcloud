package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

// FlexBool is a bool that can unmarshal from a JSON bool, string, or number.
// Strings like "Enabled", "true", "1" are treated as true.
// Strings like "Suspended", "Disabled", "false", "0", "" are treated as false.
// Numbers: 0 is false, non-zero is true.
type FlexBool bool

func (f *FlexBool) UnmarshalJSON(data []byte) error {
	// Try bool first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*f = FlexBool(b)
		return nil
	}

	// Try number
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexBool(n != 0)
		return nil
	}

	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		switch strings.ToLower(s) {
		case "enabled", "true", "1", "yes", "on":
			*f = true
		default:
			*f = false
		}
		return nil
	}

	// Default to false for null or unparseable
	*f = false
	return nil
}

// FlexInt64 is an int64 that can unmarshal from either a JSON string or number
type FlexInt64 int64

func (f *FlexInt64) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int64 first
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexInt64(i)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Parse string to int64
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*f = FlexInt64(i)
	return nil
}

// BucketReplicationConfig represents replication settings for a bucket
type BucketReplicationConfig struct {
	ReplicationType string `json:"replicationType,omitempty"`
	Region          string `json:"region,omitempty"`
	AZ              string `json:"az,omitempty"`
	Tag             string `json:"tag,omitempty"`
}

// BucketCreateConfig represents the config object for bucket creation
type BucketCreateConfig struct {
	Quota               int64                    `json:"quota,omitempty"`
	Versioning          bool                     `json:"versioning,omitempty"`
	ObjLocking          bool                     `json:"objLocking,omitempty"`
	Replication         *BucketReplicationConfig `json:"replication,omitempty"`
	ObjLockValidityDays int64                    `json:"objLockValidityDays,omitempty"`
}

// CreateObjectStorageBucketRequest represents the request to create an object storage bucket
// Uses nested structure: {"bucket": "name", "config": {...}, "tags": {...}}
type CreateObjectStorageBucketRequest struct {
	Bucket string              `json:"bucket"`
	Config *BucketCreateConfig `json:"config,omitempty"`
	Tags   map[string]string   `json:"tags,omitempty"`
}

// ObjectStorageBucket represents an object storage bucket response from the API
type ObjectStorageBucket struct {
	Domain            string                   `json:"domain"`
	Name              string                   `json:"name"`
	CreateTime        string                   `json:"createTime"`
	State             string                   `json:"state"`
	Usage             FlexInt64                `json:"usage"`
	Objects           int                      `json:"objects"`
	ObjLocking        *FlexBool                `json:"objLocking,omitempty"`
	S3Endpoint        string                   `json:"s3Endpoint"`
	PublicEndpoint    string                   `json:"publicEndpoint"`
	ReplicationConfig *BucketReplicationConfig `json:"replicationConfig,omitempty"`
	Tags              map[string]string        `json:"tags,omitempty"`
	Versioning        *FlexBool                `json:"versioning,omitempty"`
}

// UpdateObjectStorageBucketRequest represents the request to update an object storage bucket
type UpdateObjectStorageBucketRequest struct {
	Versioning *bool             `json:"versioning,omitempty"`
	ObjLocking *bool             `json:"objLocking,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// ObjectStorageBucketListResponse represents the response for listing object storage buckets
type ObjectStorageBucketListResponse struct {
	Count int                   `json:"count"`
	Items []ObjectStorageBucket `json:"items"`
}

// CreateAccessKeyRequest represents the request to create an object storage access key
type CreateAccessKeyRequest struct {
	Expiry int64  `json:"expiry"`
	AZ     string `json:"az,omitempty"`
}

// CreateAccessKeyResponse represents the response from creating an access key
type CreateAccessKeyResponse struct {
	URL       string `json:"url"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// AccessKey represents an object storage access key from the list API
type AccessKey struct {
	AccessKey   string    `json:"accessKey"`
	AccessKeyID string    `json:"accessKeyId"`
	SecretKey   string    `json:"secretKey"`
	Expiry      FlexInt64 `json:"expiry"`
	CreateTime  FlexInt64 `json:"createTime"`
}

// AccessKeyListResponse represents the response for listing access keys
type AccessKeyListResponse struct {
	Count int         `json:"count"`
	Items []AccessKey `json:"items"`
}

// LifecycleRule represents a lifecycle rule for object storage
type LifecycleRule struct {
	ID                     string `json:"id"`
	Status                 string `json:"status"`
	Prefix                 string `json:"prefix,omitempty"`
	ExpirationDays         int    `json:"expiration_days,omitempty"`
	TransitionDays         int    `json:"transition_days,omitempty"`
	TransitionStorageClass string `json:"transition_storage_class,omitempty"`
}

// CORSRule represents CORS configuration for object storage
type CORSRule struct {
	AllowedHeaders []string `json:"allowed_headers,omitempty"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedOrigins []string `json:"allowed_origins"`
	ExposeHeaders  []string `json:"expose_headers,omitempty"`
	MaxAgeSeconds  int      `json:"max_age_seconds,omitempty"`
}
