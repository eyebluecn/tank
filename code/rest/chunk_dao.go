package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/uuid"
	"time"
)

// UploadSessionDao handles database operations for UploadSession
type UploadSessionDao struct {
	BaseDao
}

// FindByUuid finds an upload session by uuid
func (this *UploadSessionDao) FindByUuid(sessionUuid string) *UploadSession {
	var entity = &UploadSession{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", sessionUuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// CheckByUuid checks if session exists, panics if not found
func (this *UploadSessionDao) CheckByUuid(sessionUuid string) *UploadSession {
	entity := this.FindByUuid(sessionUuid)
	if entity == nil {
		panic(result.NotFound("upload session not found with uuid = %s", sessionUuid))
	}
	return entity
}

// FindByUserUuidAndFileMd5 finds session by user and file MD5 (for quick upload)
func (this *UploadSessionDao) FindByUserUuidAndFileMd5(userUuid string, fileMd5 string) *UploadSession {
	var entity = &UploadSession{}
	db := core.CONTEXT.GetDB().Where("user_uuid = ? AND file_md5 = ? AND status = ?",
		userUuid, fileMd5, UPLOAD_SESSION_STATUS_COMPLETED).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// FindBySpaceUuidAndFileMd5 finds completed session by space and file MD5
func (this *UploadSessionDao) FindBySpaceUuidAndFileMd5(spaceUuid string, fileMd5 string) *UploadSession {
	var entity = &UploadSession{}
	db := core.CONTEXT.GetDB().Where("space_uuid = ? AND file_md5 = ? AND status = ?",
		spaceUuid, fileMd5, UPLOAD_SESSION_STATUS_COMPLETED).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// Create creates a new upload session
func (this *UploadSessionDao) Create(session *UploadSession) *UploadSession {
	timeUUID, _ := uuid.NewV4()
	session.Uuid = timeUUID.String()
	session.CreateTime = time.Now()
	session.UpdateTime = time.Now()
	session.Sort = time.Now().UnixNano() / 1e6

	db := core.CONTEXT.GetDB().Create(session)
	this.PanicError(db.Error)
	return session
}

// Save updates an upload session
func (this *UploadSessionDao) Save(session *UploadSession) *UploadSession {
	session.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(session)
	this.PanicError(db.Error)
	return session
}

// Delete deletes an upload session
func (this *UploadSessionDao) Delete(session *UploadSession) {
	db := core.CONTEXT.GetDB().Delete(session)
	this.PanicError(db.Error)
}

// DeleteByUuid deletes session by uuid
func (this *UploadSessionDao) DeleteByUuid(sessionUuid string) {
	db := core.CONTEXT.GetDB().Where("uuid = ?", sessionUuid).Delete(UploadSession{})
	this.PanicError(db.Error)
}

// FindExpiredSessions finds all expired sessions
func (this *UploadSessionDao) FindExpiredSessions() []*UploadSession {
	var sessions []*UploadSession
	db := core.CONTEXT.GetDB().Where("expire_time < ? AND status = ?",
		time.Now(), UPLOAD_SESSION_STATUS_UPLOADING).Find(&sessions)
	this.PanicError(db.Error)
	return sessions
}

// FindByUserUuid finds all sessions for a user
func (this *UploadSessionDao) FindByUserUuid(userUuid string) []*UploadSession {
	var sessions []*UploadSession
	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).Find(&sessions)
	this.PanicError(db.Error)
	return sessions
}

// DeleteByUserUuid deletes all sessions for a user
func (this *UploadSessionDao) DeleteByUserUuid(userUuid string) {
	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).Delete(UploadSession{})
	this.PanicError(db.Error)
}

// UploadChunkDao handles database operations for UploadChunk
type UploadChunkDao struct {
	BaseDao
}

// FindByUuid finds a chunk by uuid
func (this *UploadChunkDao) FindByUuid(chunkUuid string) *UploadChunk {
	var entity = &UploadChunk{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", chunkUuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// FindBySessionUuidAndChunkIndex finds a chunk by session and index
func (this *UploadChunkDao) FindBySessionUuidAndChunkIndex(sessionUuid string, chunkIndex int) *UploadChunk {
	var entity = &UploadChunk{}
	db := core.CONTEXT.GetDB().Where("session_uuid = ? AND chunk_index = ?", sessionUuid, chunkIndex).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// FindBySessionUuid finds all chunks for a session
func (this *UploadChunkDao) FindBySessionUuid(sessionUuid string) []*UploadChunk {
	var chunks []*UploadChunk
	db := core.CONTEXT.GetDB().Where("session_uuid = ?", sessionUuid).Order("chunk_index asc").Find(&chunks)
	this.PanicError(db.Error)
	return chunks
}

// FindUploadedBySessionUuid finds all uploaded chunks for a session
func (this *UploadChunkDao) FindUploadedBySessionUuid(sessionUuid string) []*UploadChunk {
	var chunks []*UploadChunk
	db := core.CONTEXT.GetDB().Where("session_uuid = ? AND uploaded = ?", sessionUuid, true).Order("chunk_index asc").Find(&chunks)
	this.PanicError(db.Error)
	return chunks
}

// CountUploadedBySessionUuid counts uploaded chunks for a session
func (this *UploadChunkDao) CountUploadedBySessionUuid(sessionUuid string) int64 {
	var count int64
	db := core.CONTEXT.GetDB().Model(&UploadChunk{}).Where("session_uuid = ? AND uploaded = ?", sessionUuid, true).Count(&count)
	this.PanicError(db.Error)
	return count
}

// Create creates a new chunk record
func (this *UploadChunkDao) Create(chunk *UploadChunk) *UploadChunk {
	timeUUID, _ := uuid.NewV4()
	chunk.Uuid = timeUUID.String()
	chunk.CreateTime = time.Now()
	chunk.UpdateTime = time.Now()
	chunk.Sort = time.Now().UnixNano() / 1e6

	db := core.CONTEXT.GetDB().Create(chunk)
	this.PanicError(db.Error)
	return chunk
}

// Save updates a chunk record
func (this *UploadChunkDao) Save(chunk *UploadChunk) *UploadChunk {
	chunk.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(chunk)
	this.PanicError(db.Error)
	return chunk
}

// Delete deletes a chunk record
func (this *UploadChunkDao) Delete(chunk *UploadChunk) {
	db := core.CONTEXT.GetDB().Delete(chunk)
	this.PanicError(db.Error)
}

// DeleteBySessionUuid deletes all chunks for a session
func (this *UploadChunkDao) DeleteBySessionUuid(sessionUuid string) {
	db := core.CONTEXT.GetDB().Where("session_uuid = ?", sessionUuid).Delete(UploadChunk{})
	this.PanicError(db.Error)
}
