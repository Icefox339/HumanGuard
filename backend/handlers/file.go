// backend/handlers/file.go
package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

var (
	errFileTooLarge    = errors.New("file exceeds maximum size")
	errUnsupportedType = errors.New("unsupported file type")
)

var allowedTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
	"text/plain":      true,
	"text/csv":        true,
	"application/zip": true,
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

	mr, err := r.MultipartReader()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart request"})
		return
	}

	uploadID := uuid.New().String()
	h.mu.Lock()
	h.progress[uploadID] = &UploadProgress{
		UploadID:   uploadID,
		TotalBytes: contentLength,
	}
	h.mu.Unlock()

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

		if part.FormName() == "file" {
			filename := part.FileName()
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
						pw.Write(buf[:n])
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
				h.s3.Delete(path)
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
	if err != nil {
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
	io.Copy(w, reader)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	fileRecord, err := h.store.GetFile(r.Context(), fileID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.s3.Delete(fileRecord.Path)
	h.store.DeleteFile(r.Context(), fileID)

	w.WriteHeader(http.StatusNoContent)
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

	b := make([]byte, 32)
	rand.Read(b)
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
		"token":     token,
		"share_url": "http://localhost:8080/api/files/share/" + token,
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
	io.Copy(w, reader)
}

func (h *FileHandler) UploadProgressWS(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")
	if uploadID == "" {
		http.Error(w, "upload_id required", http.StatusBadRequest)
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
			conn.WriteJSON(UploadProgress{UploadID: uploadID, Completed: true, Percentage: 100})
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
	json.NewEncoder(w).Encode(data)
}