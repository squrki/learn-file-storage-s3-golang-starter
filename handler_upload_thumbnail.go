package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't Parse Multipart Form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	// file, _, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't Get FormFile", err)
		return
	}
	contentType := header.Header.Get("Content-Type")

	// imageData, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Couldn't Read File", err)
	// 	return
	// }
	// imageDataBase64 := base64.StdEncoding.EncodeToString(imageData)
	// dataURL := "data:image/png;base64," + imageDataBase64

	filename := filepath.Join(cfg.assetsRoot, videoID.String()+"."+contentType[len(contentType)-3:])
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
	// videoThumbnails[videoID] = thumbnail{
	// 	data:      imageData,
	// 	mediaType: contentType,
	// }
	url := "http://localhost:8091/assets/" + videoID.String() + "." + contentType[len(contentType)-3:]
	videoData.ThumbnailURL = &url
	// videoData.ThumbnailURL = &dataURL

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Update Video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
