package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerThumbnailGet(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}
	// tn, ok := videoThumbnails[videoID]
	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get video metadata", err)
		return
	}
	tn := *videoData.ThumbnailURL
	u, _ := url.Parse(tn)
	parts := strings.SplitN(u.Opaque, ",", 2)
	mimeTypeAndEncoding := parts[0]
	encodedData := parts[1]
	// if !ok {
	// 	respondWithError(w, http.StatusNotFound, "Thumbnail not found", nil)
	// 	return
	// }
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		return
	}
	w.Header().Set("Content-Type", mimeTypeAndEncoding)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(tn)))

	_, err = w.Write(decodedData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error writing response", err)
		return
	}
}
