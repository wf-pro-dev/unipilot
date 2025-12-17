package errors

type ErrorCode string

const (
	// ============================================
	// INITIALIZATION ERRORS (init-xxxx)
	// ============================================

	InitDatabaseNotInitialized ErrorCode = "init-0001" // database not initialized
	InitUserNotAuthenticated   ErrorCode = "init-0002" // user not authenticated

	// ============================================
	// DATABASE ERRORS (db-xxxx)
	// ============================================
	DBConnectionFailed    ErrorCode = "db-0001" // database connection failed
	DBRecordNotFound      ErrorCode = "db-0002" // database record not found
	DBQueryFailed         ErrorCode = "db-0003" // database query failed (general - covers SELECT/INSERT/UPDATE/DELETE)
	DBTransactionFailed   ErrorCode = "db-0004" // database transaction failed
	DBConstraintViolation ErrorCode = "db-0005" // database constraint violation (covers conflicts/duplicates)

	// ============================================
	// VALIDATION ERRORS (val-xxxx)
	// ============================================
	ValidationRequired  ErrorCode = "val-0001" // validation required
	ValidationInvalid   ErrorCode = "val-0002" // validation invalid
	NonParsableSchedule ErrorCode = "val-0100" // non parsable schedule

	// ============================================
	// NETWORK ERRORS (net-xxxx)
	// ============================================
	NetworkConnectionFailed ErrorCode = "net-0001" // network connection failed
	NetworkOffline          ErrorCode = "net-0002" // network offline
	APIStartFailed          ErrorCode = "net-0003" // API start failed
	GRPCFailed              ErrorCode = "net-0004" // gRPC failed
	RedisFailed             ErrorCode = "net-0005" // Redis failed
	QdrantFailed            ErrorCode = "net-0006" // Qdrant failed
	GeminiFailed            ErrorCode = "net-0007" // Gemini failed
	SlowRequest             ErrorCode = "net-0008" // slow request

	// Network - gRPC specific
	GRPCClientFailed       ErrorCode = "net-0101" // gRPC client failed
	GRPCUnreachable        ErrorCode = "net-0102" // gRPC unreachable
	GRPCCloseFailed        ErrorCode = "net-0103" // gRPC close failed
	GRPCNotificationFailed ErrorCode = "net-0104" // gRPC notification failed

	// Network - Redis specific
	RedisUnreachable ErrorCode = "net-0201" // Redis unreachable
	RedisCloseFailed ErrorCode = "net-0202" // Redis close failed

	// Network - Qdrant specific
	QdrantClientFailed ErrorCode = "net-0301" // Qdrant client failed
	QdrantUnreachable  ErrorCode = "net-0302" // Qdrant unreachable
	QdrantCloseFailed  ErrorCode = "net-0303" // Qdrant close failed
	// ============================================
	// CACHE ERRORS (cache-xxxx)
	// ============================================
	CacheConnectionFailed ErrorCode = "cache-0001" // cache connection failed
	CacheOperationFailed  ErrorCode = "cache-0002" // cache operation failed (general - covers GET/SET/DELETE/key not found)

	// ============================================
	// STORAGE ERRORS (storage-xxxx)
	// ============================================
	StorageClientFailed   ErrorCode = "storage-0001" // storage client failed
	StorageFileNotFound   ErrorCode = "storage-0002" // storage file not found
	StorageApiFailed      ErrorCode = "storage-0003" // storage API failed
	StorageUploadFailed   ErrorCode = "storage-0004" // storage upload failed (general - covers copy operations)
	StorageDownloadFailed ErrorCode = "storage-0005" // storage download failed
	StorageCopyFailed     ErrorCode = "storage-0006" // storage copy failed
	StorageDeleteFailed   ErrorCode = "storage-0007" // storage delete failed
	StorageQuotaExceeded  ErrorCode = "storage-0008" // storage quota exceeded

	// ============================================
	// FILE SYSTEM ERRORS (fs-xxxx)
	// ============================================
	FSPathNotFound         ErrorCode = "fs-0001" // path not found
	FSFileNotFound         ErrorCode = "fs-0002" // local file not found
	FSOpenFailed           ErrorCode = "fs-0003" // file open failed (covers read operations)
	FSCreateFailed         ErrorCode = "fs-0004" // file create failed
	FSWriteFailed          ErrorCode = "fs-0005" // file write failed
	FSDeleteFailed         ErrorCode = "fs-0006" // file delete failed
	FSDirCreateFailed      ErrorCode = "fs-0007" // directory create failed
	FSFileTooLarge         ErrorCode = "fs-0008" // file size exceeded
	FSStreamFailed         ErrorCode = "fs-0009" // file stream failed
	FSFileTypeNotSupported ErrorCode = "fs-0010" // file type not supported

	// ============================================
	// AUTHENTICATION ERRORS (auth-xxxx)
	// ============================================
	AuthUnauthorized    ErrorCode = "auth-0001" // unauthorized
	AuthTokenInvalid    ErrorCode = "auth-0002" // token invalid
	AuthTokenExpired    ErrorCode = "auth-0003" // token expired
	AuthTokenGeneration ErrorCode = "auth-0004" // Token generation failed
	AuthForbidden       ErrorCode = "auth-0005" // forbidden access

	// ============================================
	// REQUEST ERRORS (req-xxxx)
	// ============================================
	ReqBodyInvalid  ErrorCode = "req-0001" // invalid request body (used 15+ times)
	ReqParamMissing ErrorCode = "req-0002" // missing parameter (covers metadata, file, query params)
	ReqParamInvalid ErrorCode = "req-0003" // invalid parameter

	// ============================================
	// PROCESSING ERRORS (proc-xxxx)
	// ============================================
	ProcJSONMarshalFailed    ErrorCode = "proc-0001" // JSON marshal failed (used 63+ times)
	ProcJSONUnmarshalFailed  ErrorCode = "proc-0002" // JSON unmarshal failed (used 63+ times)
	ProcDataConversionFailed ErrorCode = "proc-0003" // data conversion failed (used 59+ times - covers strconv operations)
	ProcDataProcessingFailed ErrorCode = "proc-0004" // data processing failed (used 36+ times - covers ToMap operations)

	// ============================================
	// CONFIGURATION ERRORS (config-xxxx)
	// ============================================
	ConfigEnvVarNotFound ErrorCode = "config-0001" // environment variable not found (covers all config errors)

	// ============================================
	// QDRANT ERRORS (qdrant-xxxx)
	// ============================================
	QdrantVectorsError          ErrorCode = "qdrant-0001" // Qdrant vectors error
	QdrantTextError             ErrorCode = "qdrant-0002" // Qdrant text error
	QdrantEmbeddingError        ErrorCode = "qdrant-0003" // Qdrant embedding error
	QdrantUpsertError           ErrorCode = "qdrant-0004" // Qdrant upsert error
	QdrantCollectionNotFound    ErrorCode = "qdrant-0005" // Qdrant collection not found
	QdrantListCollectionsError  ErrorCode = "qdrant-0006" // Qdrant list collections error
	QdrantCreateCollectionError ErrorCode = "qdrant-0007" // Qdrant create collection error
	QdrantDeleteCollectionError ErrorCode = "qdrant-0008" // Qdrant delete collection error
	QdrantDeletePointError      ErrorCode = "qdrant-0009" // Qdrant delete point error
	QdrantDeletePointsError     ErrorCode = "qdrant-0010" // Qdrant delete points error

	// ============================================
	// API/SERVER ERRORS (api-xxxx)
	// ============================================
	APIServerStartFailed ErrorCode = "api-0001" // server start failed
	APIUnhealthy         ErrorCode = "api-0002" // health check failed

	// ============================================
	// INTERNAL ERRORS (internal-xxxx)
	// ============================================
	InternalError ErrorCode = "internal-0001" // internal error (fallback for unknown errors)

	// ============================================
	// SYSTEM ERRORS (sys-xxxx)
	// ============================================
	SysExecFailed ErrorCode = "sys-0001" // system execute failed
)
