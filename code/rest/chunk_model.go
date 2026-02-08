package rest

import (
	"time"
)

const (
	// Upload session status
	UPLOAD_SESSION_STATUS_UPLOADING = "UPLOADING"
	UPLOAD_SESSION_STATUS_MERGING   = "MERGING"
	UPLOAD_SESSION_STATUS_COMPLETED = "COMPLETED"
	UPLOAD_SESSION_STATUS_ERROR     = "ERROR"

	// Chunk directory name
	CHUNK_DIR = "chunk"

	// Default chunk size: 50MB
	DEFAULT_CHUNK_SIZE = 50 * 1024 * 1024

	// Max chunk size: 100MB
	MAX_CHUNK_SIZE = 100 * 1024 * 1024

	// Min chunk size: 5MB (aligned with AWS S3 standard)
	MIN_CHUNK_SIZE = 5 * 1024 * 1024
)

// UploadSession represents a chunked upload session
type UploadSession struct {
	Uuid        string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort        int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime  time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime  time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	UserUuid    string    `json:"userUuid" gorm:"type:char(36) not null;index:idx_upload_session_uu"`
	SpaceUuid   string    `json:"spaceUuid" gorm:"type:char(36) not null;index:idx_upload_session_su"`
	MatterUuid  string    `json:"matterUuid" gorm:"type:char(36)"`                                    // Final matter uuid after merge
	Puuid       string    `json:"puuid" gorm:"type:char(36) not null"`                                // Parent directory uuid
	Filename    string    `json:"filename" gorm:"type:varchar(255) not null"`                         // Target filename
	TotalSize   int64     `json:"totalSize" gorm:"type:bigint(20) not null;default:0"`                // Total file size
	ChunkSize   int64     `json:"chunkSize" gorm:"type:bigint(20) not null;default:52428800"`         // Size of each chunk (default 50MB)
	TotalChunks int       `json:"totalChunks" gorm:"type:int not null;default:0"`                     // Total number of chunks
	Privacy     bool      `json:"privacy" gorm:"type:tinyint(1) not null;default:1"`                  // Privacy flag
	FileMd5     string    `json:"fileMd5" gorm:"type:varchar(45)"`                                    // MD5 of the whole file
	Status      string    `json:"status" gorm:"type:varchar(20) not null;default:'UPLOADING'"`        // Session status
	ExpireTime  time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"` // Expiration time
}

// UploadChunk represents a single chunk in an upload session
type UploadChunk struct {
	Uuid          string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort          int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime    time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime    time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	SessionUuid   string    `json:"sessionUuid" gorm:"type:char(36) not null;index:idx_upload_chunk_su"` // Reference to UploadSession
	ChunkIndex    int       `json:"chunkIndex" gorm:"type:int not null"`                                 // Chunk sequence number (0-based)
	ChunkSize     int64     `json:"chunkSize" gorm:"type:bigint(20) not null;default:0"`                 // Actual size of this chunk
	ChunkMd5      string    `json:"chunkMd5" gorm:"type:varchar(45)"`                                    // MD5 of this chunk
	ChunkFilename string    `json:"chunkFilename" gorm:"type:varchar(255) not null"`                     // Temp filename for this chunk
	Uploaded      bool      `json:"uploaded" gorm:"type:tinyint(1) not null;default:0"`                  // Whether this chunk is uploaded
}

// Get the chunk temp directory for a space
func GetSpaceChunkDir(spaceName string) string {
	return GetUserSpaceRootDir(spaceName) + "/" + CHUNK_DIR
}

// Get the session chunk directory
func GetSessionChunkDir(spaceName string, sessionUuid string) string {
	return GetSpaceChunkDir(spaceName) + "/" + sessionUuid
}

// UploadSessionInfo is the response structure for session queries
type UploadSessionInfo struct {
	Session        *UploadSession `json:"session"`
	UploadedChunks []int          `json:"uploadedChunks"` // List of uploaded chunk indices
	MissingChunks  []int          `json:"missingChunks"`  // List of missing chunk indices
}
