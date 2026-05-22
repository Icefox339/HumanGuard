package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"humanguard/auth"
	"humanguard/storage"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// var (
// 	errFileTooLarge    = errors.New("file exceeds maximum size")
// 	errUnsupportedType = errors.New("unsupported file type")
// )

var allowedTypes = map[string]bool{
	"image/jpeg":       true,
	"image/png":        true,
	"image/gif":        true,
	"image/webp":       true,
	"application/pdf":  true,
	"text/plain":       true,
	"text/csv":         true,
	"application/zip":  true,
	"application/json": true,
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type FileHandler struct {
	store    storage.Storage
	s3       storage.S3Client
	progress map[string]*UploadProgress
	mu       sync.RWMutex
}

type UploadProgress struct {
	UploadID   string `json:"upload_id"`
	UserID     string `json:"-"`
	BytesDone  int64  `json:"bytes_done"`
	TotalBytes int64  `json:"total_bytes"`
	Percentage int    `json:"percentage"`
	Completed  bool   `json:"completed"`
}

func NewFileHandler(store storage.Storage, s3 storage.S3Client) *FileHandler {
	return &FileHandler{
		store:    store,
		s3:       s3,
		progress: make(map[string]*UploadProgress),
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<30)

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	contentLength := r.ContentLength
	if contentLength <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content-length required"})
		return
	}

	uploadID := r.URL.Query().Get("upload_id")
	if uploadID == "" {
		uploadID = uuid.New().String()
	}

	h.mu.Lock()
	h.progress[uploadID] = &UploadProgress{
		UploadID:   uploadID,
		UserID:     userID,
		TotalBytes: contentLength,
	}
	h.mu.Unlock()

	mr, err := r.MultipartReader()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart request"})
		return
	}

	var fileRecord *storage.FileRecord
	var bytesRead int64

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		formName := part.FormName()
		fileName := part.FileName()

		if formName == "file" {
			filename := fileName
			mimeType := part.Header.Get("Content-Type")

			if !allowedTypes[mimeType] {
				writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "unsupported file type"})
				return
			}

			ext := filepath.Ext(filename)
			safeName := uuid.New().String() + ext
			path := fmt.Sprintf("%s/%s/%s", userID, time.Now().Format("2006/01/02"), safeName)

			hasher := sha256.New()
			buf := make([]byte, 32*1024)
			pr, pw := io.Pipe()

			go func() {
				defer pw.Close()
				for {
					n, readErr := part.Read(buf)
					if n > 0 {
						bytesRead += int64(n)
						if _, err := pw.Write(buf[:n]); err != nil {
							return
						}
						h.mu.Lock()
						if p, ok := h.progress[uploadID]; ok {
							p.BytesDone = bytesRead
							p.Percentage = int(bytesRead * 100 / contentLength)
						}
						h.mu.Unlock()
					}
					if readErr != nil {
						break
					}
				}
			}()

			teeReader := io.TeeReader(pr, hasher)
			size, err := h.s3.Save(path, teeReader)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save file"})
				return
			}

			h.mu.Lock()
			if p, ok := h.progress[uploadID]; ok {
				p.Completed = true
				p.Percentage = 100
			}
			h.mu.Unlock()

			fileRecord = &storage.FileRecord{
				ID:           uuid.New().String(),
				UserID:       userID,
				Name:         safeName,
				OriginalName: filename,
				Size:         size,
				MimeType:     mimeType,
				Hash:         hex.EncodeToString(hasher.Sum(nil)),
				Path:         path,
				CreatedAt:    time.Now(),
			}

			if err := h.store.CreateFile(r.Context(), fileRecord); err != nil {
				_ = h.s3.Delete(path)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save file metadata"})
				return
			}

			break
		}
	}

	if fileRecord == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no file provided"})
		return
	}

	writeJSON(w, http.StatusCreated, fileRecord)
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	fileRecord, err := h.store.GetFile(r.Context(), fileID)
	if err != nil || fileRecord.UserID != auth.GetUserID(r.Context()) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	reader, err := h.s3.Get(fileRecord.Path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", fileRecord.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileRecord.OriginalName+"\"")
	if _, err := io.Copy(w, reader); err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	fileRecord, err := h.store.GetFile(r.Context(), fileID)
	if err != nil || fileRecord.UserID != auth.GetUserID(r.Context()) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	originalName := fileRecord.OriginalName

	if err := h.s3.Delete(fileRecord.Path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete file from storage"})
		return
	}
	if err := h.store.DeleteFile(r.Context(), fileID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete file metadata"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":       "file deleted successfully",
		"file_id":       fileID,
		"original_name": originalName,
	})
}

func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	files, err := h.store.ListUserFiles(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list files"})
		return
	}

	if files == nil {
		files = []*storage.FileRecord{}
	}

	writeJSON(w, http.StatusOK, files)
}

func (h *FileHandler) CreateShare(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FileID    string `json:"file_id"`
		ExpiresIn int    `json:"expires_in_hours"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	fileRecord, err := h.store.GetFile(r.Context(), req.FileID)
	if err != nil || fileRecord.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}
	token := hex.EncodeToString(b)

	share := &storage.ShareRecord{
		FileID:    req.FileID,
		Token:     token,
		SharedBy:  userID,
		CreatedAt: time.Now(),
	}

	if req.ExpiresIn > 0 {
		share.ExpiresAt = time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
	}

	if _, err := h.store.CreateShare(r.Context(), share); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create share"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func (h *FileHandler) GetByShareToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	fileRecord, err := h.store.GetFileByShareToken(r.Context(), token)
	if err != nil {
		http.Error(w, "not found or expired", http.StatusNotFound)
		return
	}

	reader, err := h.s3.Get(fileRecord.Path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", fileRecord.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileRecord.OriginalName+"\"")
	if _, err := io.Copy(w, reader); err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) UploadProgressWS(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uploadID := r.URL.Query().Get("upload_id")
	if uploadID == "" {
		http.Error(w, "upload_id required", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	progress, exists := h.progress[uploadID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	if progress.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		p, ok := h.progress[uploadID]
		h.mu.RUnlock()

		if !ok {
			if err := conn.WriteJSON(UploadProgress{UploadID: uploadID, Completed: true, Percentage: 100}); err != nil {
				return
			}
			return
		}

		if err := conn.WriteJSON(p); err != nil {
			return
		}

		if p.Completed {
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
