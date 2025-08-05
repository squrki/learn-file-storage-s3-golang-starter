package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 1 << 30
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't Parse Multipart Form", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't Get FormFile", err)
		return
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType("video/mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't Parse Media Type", err)
		return
	}
	if mediaType != contentType {
		respondWithError(w, http.StatusBadRequest, "Media Type mismatch", err)
		return
	}

	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't write temp file", err)
		return
	}
	defer os.Remove("tubely-upload.mp4")
	defer tempFile.Close()
	io.Copy(tempFile, file)

	tempFile.Seek(0, io.SeekStart)

	cfg.s3Client.PutObject(r.Context())

	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	nameBase64 := base64.RawURLEncoding.EncodeToString(randomBytes)

	filename := filepath.Join(cfg.assetsRoot, nameBase64+"."+contentType[len(contentType)-3:])
	assetFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}
	io.Copy(assetFile, file)

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get video metadata", err)
		return
	}
	url := "http://localhost:8091/assets/" + nameBase64 + "." + contentType[len(contentType)-3:]
	videoData.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Update Video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
