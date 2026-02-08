package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
)

// ChunkController handles chunked upload API endpoints
type ChunkController struct {
	BaseController
	chunkService     *ChunkService
	matterDao        *MatterDao
	uploadSessionDao *UploadSessionDao
	uploadChunkDao   *UploadChunkDao
}

func (this *ChunkController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.chunkService)
	if b, ok := b.(*ChunkService); ok {
		this.chunkService = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadSessionDao)
	if b, ok := b.(*UploadSessionDao); ok {
		this.uploadSessionDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadChunkDao)
	if b, ok := b.(*UploadChunkDao); ok {
		this.uploadChunkDao = b
	}
}

func (this *ChunkController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {
	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	// Chunked upload endpoints
	routeMap["/api/chunk/create/session"] = this.Wrap(this.CreateSession, USER_ROLE_USER)
	routeMap["/api/chunk/upload"] = this.Wrap(this.Upload, USER_ROLE_USER)
	routeMap["/api/chunk/merge"] = this.Wrap(this.Merge, USER_ROLE_USER)
	routeMap["/api/chunk/session/info"] = this.Wrap(this.SessionInfo, USER_ROLE_USER)
	routeMap["/api/chunk/cancel"] = this.Wrap(this.Cancel, USER_ROLE_USER)
	routeMap["/api/chunk/clean/expired"] = this.Wrap(this.CleanExpired, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

// CreateSession creates a new chunked upload session
// Parameters:
//   - puuid: parent directory uuid
//   - filename: target filename
//   - totalSize: total file size in bytes
//   - chunkSize: size of each chunk (optional, default 5MB)
//   - privacy: whether the file is private (optional, default true)
//   - fileMd5: MD5 of the whole file (optional, for verification)
//   - spaceUuid: space uuid (optional, default user's private space)
func (this *ChunkController) CreateSession(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	puuid := util.ExtractRequestString(request, "puuid")
	filename := util.ExtractRequestString(request, "filename")
	totalSize := util.ExtractRequestInt64(request, "totalSize")
	chunkSize := util.ExtractRequestOptionalInt64(request, "chunkSize", DEFAULT_CHUNK_SIZE)
	privacy := util.ExtractRequestOptionalBool(request, "privacy", true)
	fileMd5 := util.ExtractRequestOptionalString(request, "fileMd5", "")

	user := this.checkUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	dirMatter := this.matterDao.CheckWithRootByUuid(puuid, space)

	session := this.chunkService.CreateSession(request, user, space, dirMatter, filename, totalSize, chunkSize, privacy, fileMd5)

	return this.Success(session)
}

// Upload uploads a single chunk
// Parameters:
//   - sessionUuid: upload session uuid
//   - chunkIndex: chunk index (0-based)
//   - file: chunk file data
//   - chunkMd5: MD5 of the chunk (optional, for verification)
//   - spaceUuid: space uuid (optional)
func (this *ChunkController) Upload(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	sessionUuid := util.ExtractRequestString(request, "sessionUuid")
	chunkIndex := util.ExtractRequestInt(request, "chunkIndex")
	chunkMd5 := util.ExtractRequestOptionalString(request, "chunkMd5", "")

	user := this.checkUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	file, _, err := request.FormFile("file")
	if err != nil {
		panic(result.BadRequest("file is required: %s", err.Error()))
	}
	defer file.Close()

	chunk := this.chunkService.UploadChunk(request, sessionUuid, chunkIndex, file, chunkMd5, user, space)

	return this.Success(chunk)
}

// Merge merges all chunks into the final file
// Parameters:
//   - sessionUuid: upload session uuid
//   - spaceUuid: space uuid (optional)
func (this *ChunkController) Merge(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	sessionUuid := util.ExtractRequestString(request, "sessionUuid")

	user := this.checkUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	matter := this.chunkService.MergeChunks(request, sessionUuid, user, space)

	return this.Success(matter)
}

// SessionInfo gets the upload session info including uploaded chunks
// Parameters:
//   - sessionUuid: upload session uuid
func (this *ChunkController) SessionInfo(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	sessionUuid := util.ExtractRequestString(request, "sessionUuid")

	user := this.checkUser(request)

	sessionInfo := this.chunkService.GetSessionInfo(sessionUuid, user)

	return this.Success(sessionInfo)
}

// Cancel cancels an upload session
// Parameters:
//   - sessionUuid: upload session uuid
//   - spaceUuid: space uuid (optional)
func (this *ChunkController) Cancel(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	sessionUuid := util.ExtractRequestString(request, "sessionUuid")

	user := this.checkUser(request)
	spaceUuid := util.ExtractRequestOptionalString(request, "spaceUuid", user.SpaceUuid)
	space := this.spaceService.CheckWritableByUuid(request, user, spaceUuid)

	this.chunkService.CancelSession(sessionUuid, user, space)

	return this.Success("OK")
}

// CleanExpired cleans up expired upload sessions (admin only)
func (this *ChunkController) CleanExpired(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	this.chunkService.CleanExpiredSessions()
	return this.Success("OK")
}
