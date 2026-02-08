package rest

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
)

// ChunkService handles chunked upload operations
//@Service
type ChunkService struct {
	BaseBean
	uploadSessionDao *UploadSessionDao
	uploadChunkDao   *UploadChunkDao
	matterDao        *MatterDao
	matterService    *MatterService
	spaceDao         *SpaceDao
	userService      *UserService
	// chunkLocker is used to prevent concurrent uploads of the same chunk
	chunkLocker sync.Map
	// mergeLocker is used to prevent concurrent merge of the same session
	mergeLocker sync.Map
}

func (this *ChunkService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.uploadSessionDao)
	if b, ok := b.(*UploadSessionDao); ok {
		this.uploadSessionDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadChunkDao)
	if b, ok := b.(*UploadChunkDao); ok {
		this.uploadChunkDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
	}
}

// CreateSession creates a new upload session
func (this *ChunkService) CreateSession(
	request *http.Request,
	user *User,
	space *Space,
	dirMatter *Matter,
	filename string,
	totalSize int64,
	chunkSize int64,
	privacy bool,
	fileMd5 string,
) *UploadSession {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil."))
	}

	if dirMatter.Deleted {
		panic(result.BadRequest("Dir has been deleted. Cannot upload under it."))
	}

	// Validate filename
	if len(filename) > MATTER_NAME_MAX_LENGTH {
		panic(result.BadRequestI18n(request, i18n.MatterNameLengthExceedLimit, len(filename), MATTER_NAME_MAX_LENGTH))
	}
	CheckMatterName(request, filename)

	// Validate chunk size
	if chunkSize < MIN_CHUNK_SIZE {
		chunkSize = MIN_CHUNK_SIZE
	}
	if chunkSize > MAX_CHUNK_SIZE {
		chunkSize = MAX_CHUNK_SIZE
	}

	// Validate total size
	if totalSize <= 0 {
		panic(result.BadRequest("totalSize must be positive"))
	}

	// Check space size limit
	if space.SizeLimit >= 0 {
		if totalSize > space.SizeLimit {
			panic(result.BadRequestI18n(request, i18n.MatterSizeExceedLimit, util.HumanFileSize(totalSize), util.HumanFileSize(space.SizeLimit)))
		}
	}

	// Check total space limit
	if space.TotalSizeLimit >= 0 {
		if space.TotalSize+totalSize > space.TotalSizeLimit {
			panic(result.BadRequestI18n(request, i18n.MatterSizeExceedTotalLimit, util.HumanFileSize(space.TotalSize), util.HumanFileSize(space.TotalSizeLimit)))
		}
	}

	// Check if file already exists
	dbMatter := this.matterDao.FindBySpaceUuidAndPuuidAndDirAndName(space.Uuid, dirMatter.Uuid, false, filename)
	if dbMatter != nil {
		if dbMatter.Deleted {
			panic(result.BadRequestI18n(request, i18n.MatterRecycleBinExist, filename))
		} else {
			panic(result.BadRequestI18n(request, i18n.MatterExist, filename))
		}
	}

	// Calculate total chunks
	totalChunks := int((totalSize + chunkSize - 1) / chunkSize)

	// Create session
	session := &UploadSession{
		UserUuid:    user.Uuid,
		SpaceUuid:   space.Uuid,
		Puuid:       dirMatter.Uuid,
		Filename:    filename,
		TotalSize:   totalSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		Privacy:     privacy,
		FileMd5:     fileMd5,
		Status:      UPLOAD_SESSION_STATUS_UPLOADING,
		ExpireTime:  time.Now().Add(24 * time.Hour), // 24 hours expiration
	}

	session = this.uploadSessionDao.Create(session)

	// Create chunk directory
	chunkDir := GetSessionChunkDir(space.Name, session.Uuid)
	util.MakeDirAll(chunkDir)

	this.logger.Info("Created upload session %s for file %s, total chunks: %d", session.Uuid, filename, totalChunks)

	return session
}

// UploadChunk uploads a single chunk with optional MD5 verification
func (this *ChunkService) UploadChunk(
	request *http.Request,
	sessionUuid string,
	chunkIndex int,
	file io.Reader,
	clientChunkMd5 string,
	user *User,
	space *Space,
) *UploadChunk {

	// Check session
	session := this.uploadSessionDao.CheckByUuid(sessionUuid)

	// Validate session ownership
	if session.UserUuid != user.Uuid {
		panic(result.BadRequest("session does not belong to you"))
	}

	if session.Status != UPLOAD_SESSION_STATUS_UPLOADING {
		panic(result.BadRequest("session status is not uploading"))
	}

	// Check expiration
	if time.Now().After(session.ExpireTime) {
		panic(result.BadRequest("upload session has expired"))
	}

	// Validate chunk index
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		panic(result.BadRequest("invalid chunk index: %d, total chunks: %d", chunkIndex, session.TotalChunks))
	}

	// Acquire chunk-level lock to prevent concurrent upload of the same chunk
	chunkLockKey := fmt.Sprintf("%s_%d", sessionUuid, chunkIndex)
	if _, loaded := this.chunkLocker.LoadOrStore(chunkLockKey, true); loaded {
		panic(result.BadRequest("chunk %d is being uploaded by another request, please retry later", chunkIndex))
	}
	defer this.chunkLocker.Delete(chunkLockKey)

	// Check if chunk already exists
	existingChunk := this.uploadChunkDao.FindBySessionUuidAndChunkIndex(sessionUuid, chunkIndex)
	if existingChunk != nil && existingChunk.Uploaded {
		this.logger.Info("Chunk %d already uploaded for session %s", chunkIndex, sessionUuid)
		return existingChunk
	}

	// Prepare chunk file path
	chunkFilename := fmt.Sprintf("chunk_%d", chunkIndex)
	chunkDir := GetSessionChunkDir(space.Name, sessionUuid)
	chunkFilePath := chunkDir + "/" + chunkFilename

	// Write chunk to disk (use 0600 for security - only owner can read/write)
	destFile, err := os.OpenFile(chunkFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(result.BadRequest("failed to create chunk file: %s", err.Error()))
	}

	// Calculate MD5 while writing
	hash := md5.New()
	multiWriter := io.MultiWriter(destFile, hash)

	chunkSize, err := io.Copy(multiWriter, file)
	destFile.Close()

	if err != nil {
		os.Remove(chunkFilePath)
		panic(result.BadRequest("failed to write chunk: %s", err.Error()))
	}

	// Validate chunk size to prevent malicious oversized uploads
	isLastChunk := chunkIndex == session.TotalChunks-1
	if isLastChunk {
		// Last chunk: should be <= session.ChunkSize and match expected remainder
		expectedLastChunkSize := session.TotalSize - int64(session.TotalChunks-1)*session.ChunkSize
		if chunkSize != expectedLastChunkSize {
			os.Remove(chunkFilePath)
			panic(result.BadRequest("last chunk size mismatch: expected %d, got %d", expectedLastChunkSize, chunkSize))
		}
	} else {
		// Non-last chunks: should be exactly session.ChunkSize
		if chunkSize != session.ChunkSize {
			os.Remove(chunkFilePath)
			panic(result.BadRequest("chunk size mismatch: expected %d, got %d", session.ChunkSize, chunkSize))
		}
	}

	chunkMd5 := hex.EncodeToString(hash.Sum(nil))

	// Verify chunk MD5 if client provided one
	if clientChunkMd5 != "" && clientChunkMd5 != chunkMd5 {
		os.Remove(chunkFilePath)
		this.logger.Error("Chunk MD5 verification failed for chunk %d: client=%s, server=%s", chunkIndex, clientChunkMd5, chunkMd5)
		panic(result.BadRequest("chunk MD5 verification failed: data may be corrupted during upload"))
	}

	this.logger.Info("Uploaded chunk %d for session %s, size: %d, md5: %s", chunkIndex, sessionUuid, chunkSize, chunkMd5)

	// Create or update chunk record
	var chunk *UploadChunk
	if existingChunk != nil {
		existingChunk.ChunkSize = chunkSize
		existingChunk.ChunkMd5 = chunkMd5
		existingChunk.ChunkFilename = chunkFilename
		existingChunk.Uploaded = true
		chunk = this.uploadChunkDao.Save(existingChunk)
	} else {
		chunk = &UploadChunk{
			SessionUuid:   sessionUuid,
			ChunkIndex:    chunkIndex,
			ChunkSize:     chunkSize,
			ChunkMd5:      chunkMd5,
			ChunkFilename: chunkFilename,
			Uploaded:      true,
		}
		chunk = this.uploadChunkDao.Create(chunk)
	}

	return chunk
}

// GetSessionInfo gets the upload session info including uploaded chunks
func (this *ChunkService) GetSessionInfo(sessionUuid string, user *User) *UploadSessionInfo {
	session := this.uploadSessionDao.CheckByUuid(sessionUuid)

	// Validate ownership
	if session.UserUuid != user.Uuid {
		panic(result.BadRequest("session does not belong to you"))
	}

	// Get uploaded chunks
	uploadedChunks := this.uploadChunkDao.FindUploadedBySessionUuid(sessionUuid)

	uploadedIndices := make([]int, 0)
	uploadedMap := make(map[int]bool)
	for _, chunk := range uploadedChunks {
		uploadedIndices = append(uploadedIndices, chunk.ChunkIndex)
		uploadedMap[chunk.ChunkIndex] = true
	}

	// Calculate missing chunks
	missingIndices := make([]int, 0)
	for i := 0; i < session.TotalChunks; i++ {
		if !uploadedMap[i] {
			missingIndices = append(missingIndices, i)
		}
	}

	return &UploadSessionInfo{
		Session:        session,
		UploadedChunks: uploadedIndices,
		MissingChunks:  missingIndices,
	}
}

// MergeChunks merges all chunks into final file
func (this *ChunkService) MergeChunks(
	request *http.Request,
	sessionUuid string,
	user *User,
	space *Space,
) *Matter {

	// Use session-level lock instead of user-level lock to allow concurrent uploads of different files
	if _, loaded := this.mergeLocker.LoadOrStore(sessionUuid, true); loaded {
		panic(result.BadRequest("merge operation is already in progress for this session"))
	}
	defer this.mergeLocker.Delete(sessionUuid)

	session := this.uploadSessionDao.CheckByUuid(sessionUuid)

	// Validate ownership
	if session.UserUuid != user.Uuid {
		panic(result.BadRequest("session does not belong to you"))
	}

	// Check session expiration
	if time.Now().After(session.ExpireTime) {
		panic(result.BadRequest("upload session has expired, please start a new upload"))
	}

	if session.Status == UPLOAD_SESSION_STATUS_COMPLETED {
		// Already completed, return the matter
		if session.MatterUuid != "" {
			return this.matterDao.CheckByUuid(session.MatterUuid)
		}
		panic(result.BadRequest("session completed but matter not found"))
	}

	// Allow retry from ERROR status (e.g., previous merge failed)
	if session.Status != UPLOAD_SESSION_STATUS_UPLOADING && session.Status != UPLOAD_SESSION_STATUS_ERROR {
		panic(result.BadRequest("session status is not uploading or error, current: %s", session.Status))
	}

	// If retrying from ERROR status, reset to UPLOADING first
	if session.Status == UPLOAD_SESSION_STATUS_ERROR {
		this.logger.Info("Retrying merge for session %s from ERROR status", sessionUuid)
		session.Status = UPLOAD_SESSION_STATUS_UPLOADING
		this.uploadSessionDao.Save(session)
	}

	// Check if all chunks are uploaded
	uploadedCount := this.uploadChunkDao.CountUploadedBySessionUuid(sessionUuid)
	if int(uploadedCount) < session.TotalChunks {
		panic(result.BadRequest("not all chunks uploaded: %d/%d", uploadedCount, session.TotalChunks))
	}

	// Update status to merging
	session.Status = UPLOAD_SESSION_STATUS_MERGING
	this.uploadSessionDao.Save(session)

	// Get parent directory
	var dirMatter *Matter
	if session.Puuid == MATTER_ROOT {
		dirMatter = NewRootMatter(space)
	} else {
		dirMatter = this.matterDao.CheckByUuid(session.Puuid)
	}

	// Prepare final file path
	dirAbsolutePath := dirMatter.AbsolutePath()
	util.MakeDirAll(dirAbsolutePath)
	fileAbsolutePath := dirAbsolutePath + "/" + session.Filename

	// Check if file exists
	if util.PathExists(fileAbsolutePath) {
		this.logger.Error("%s exists, removing it.", fileAbsolutePath)
		os.Remove(fileAbsolutePath)
	}

	// Create final file (use 0644 for normal file permissions)
	destFile, err := os.OpenFile(fileAbsolutePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		session.Status = UPLOAD_SESSION_STATUS_ERROR
		this.uploadSessionDao.Save(session)
		panic(result.BadRequest("failed to create destination file: %s", err.Error()))
	}

	// Get all chunks sorted by index
	chunks := this.uploadChunkDao.FindBySessionUuid(sessionUuid)
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	})

	chunkDir := GetSessionChunkDir(space.Name, sessionUuid)

	// Merge chunks
	var totalWritten int64 = 0
	for _, chunk := range chunks {
		chunkPath := chunkDir + "/" + chunk.ChunkFilename
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			destFile.Close()
			os.Remove(fileAbsolutePath)
			session.Status = UPLOAD_SESSION_STATUS_ERROR
			this.uploadSessionDao.Save(session)
			panic(result.BadRequest("failed to open chunk %d: %s", chunk.ChunkIndex, err.Error()))
		}

		written, err := io.Copy(destFile, chunkFile)
		chunkFile.Close()

		if err != nil {
			destFile.Close()
			os.Remove(fileAbsolutePath)
			session.Status = UPLOAD_SESSION_STATUS_ERROR
			this.uploadSessionDao.Save(session)
			panic(result.BadRequest("failed to write chunk %d: %s", chunk.ChunkIndex, err.Error()))
		}

		totalWritten += written
	}

	destFile.Close()

	this.logger.Info("Merged %d chunks into file %s, total size: %d", len(chunks), session.Filename, totalWritten)

	// Verify file size
	if totalWritten != session.TotalSize {
		os.Remove(fileAbsolutePath)
		session.Status = UPLOAD_SESSION_STATUS_ERROR
		this.uploadSessionDao.Save(session)
		panic(result.BadRequest("merged file size mismatch: expected %d, got %d", session.TotalSize, totalWritten))
	}

	// Verify file MD5 if provided
	if session.FileMd5 != "" {
		mergedFile, err := os.Open(fileAbsolutePath)
		if err != nil {
			os.Remove(fileAbsolutePath)
			session.Status = UPLOAD_SESSION_STATUS_ERROR
			this.uploadSessionDao.Save(session)
			panic(result.BadRequest("failed to open merged file for MD5 verification: %s", err.Error()))
		}

		hash := md5.New()
		_, err = io.Copy(hash, mergedFile)
		mergedFile.Close()

		if err != nil {
			os.Remove(fileAbsolutePath)
			session.Status = UPLOAD_SESSION_STATUS_ERROR
			this.uploadSessionDao.Save(session)
			panic(result.BadRequest("failed to calculate MD5 of merged file: %s", err.Error()))
		}

		calculatedMd5 := hex.EncodeToString(hash.Sum(nil))
		if calculatedMd5 != session.FileMd5 {
			os.Remove(fileAbsolutePath)
			session.Status = UPLOAD_SESSION_STATUS_ERROR
			this.uploadSessionDao.Save(session)
			this.logger.Error("MD5 verification failed for session %s: expected %s, got %s", sessionUuid, session.FileMd5, calculatedMd5)
			panic(result.BadRequest("MD5 verification failed: file may be corrupted during upload"))
		}

		this.logger.Info("MD5 verification passed for session %s", sessionUuid)
	}

	// Create matter record
	dirRelativePath := dirMatter.Path
	fileRelativePath := dirRelativePath + "/" + session.Filename

	matter := &Matter{
		Puuid:     dirMatter.Uuid,
		UserUuid:  user.Uuid,
		SpaceName: space.Name,
		SpaceUuid: space.Uuid,
		Dir:       false,
		Name:      session.Filename,
		Md5:       session.FileMd5,
		Size:      totalWritten,
		Privacy:   session.Privacy,
		Path:      fileRelativePath,
		Prop:      EMPTY_JSON_MAP,
		VisitTime: time.Now(),
	}
	matter = this.matterDao.Create(matter)

	// Update session status
	session.Status = UPLOAD_SESSION_STATUS_COMPLETED
	session.MatterUuid = matter.Uuid
	this.uploadSessionDao.Save(session)

	// Clean up chunk files
	go core.RunWithRecovery(func() {
		this.cleanupChunks(space.Name, sessionUuid)
		this.matterService.ComputeRouteSize(dirMatter.Uuid, user, space)
	})

	this.logger.Info("Upload session %s completed, matter uuid: %s", sessionUuid, matter.Uuid)

	return matter
}

// cleanupChunks removes chunk files and records with improved atomicity
// First delete database records, then delete files (safer approach)
func (this *ChunkService) cleanupChunks(spaceName string, sessionUuid string) {
	chunkDir := GetSessionChunkDir(spaceName, sessionUuid)

	// First, remove chunk records from database
	// This ensures that even if file deletion fails, we don't have orphaned DB records
	this.uploadChunkDao.DeleteBySessionUuid(sessionUuid)

	// Then remove chunk directory with retry mechanism
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := os.RemoveAll(chunkDir)
		if err == nil {
			this.logger.Info("Cleaned up chunks for session %s", sessionUuid)
			return
		}

		this.logger.Error("failed to remove chunk dir %s (attempt %d/%d): %s", chunkDir, i+1, maxRetries, err.Error())

		// Wait before retry (exponential backoff)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
		}
	}

	// If all retries failed, log a critical warning
	this.logger.Error("CRITICAL: Failed to remove chunk dir %s after %d attempts. Manual cleanup may be required.", chunkDir, maxRetries)
}

// CancelSession cancels an upload session
func (this *ChunkService) CancelSession(sessionUuid string, user *User, space *Space) {
	session := this.uploadSessionDao.CheckByUuid(sessionUuid)

	// Validate ownership
	if session.UserUuid != user.Uuid {
		panic(result.BadRequest("session does not belong to you"))
	}

	// Clean up
	this.cleanupChunks(space.Name, sessionUuid)

	// Delete session
	this.uploadSessionDao.Delete(session)

	this.logger.Info("Cancelled upload session %s", sessionUuid)
}

// CleanExpiredSessions cleans up expired sessions
func (this *ChunkService) CleanExpiredSessions() {
	sessions := this.uploadSessionDao.FindExpiredSessions()

	for _, session := range sessions {
		space := this.spaceDao.FindByUuid(session.SpaceUuid)
		if space != nil {
			this.cleanupChunks(space.Name, session.Uuid)
		}
		this.uploadSessionDao.Delete(session)
		this.logger.Info("Cleaned expired session %s", session.Uuid)
	}
}
