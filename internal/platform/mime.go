package platform

import (
	"mime"
	"path/filepath"
	"strings"
)

// mediaTypes covers the containers GoEncode writes: Go's MIME table has
// few of them, and on Android, with no system table to fill in, none.
var mediaTypes = map[string]string{
	".mp4": "video/mp4", ".m4v": "video/mp4", ".mkv": "video/x-matroska",
	".webm": "video/webm", ".mov": "video/quicktime", ".avi": "video/x-msvideo",
	".ts": "video/mp2t", ".m2ts": "video/mp2t", ".mts": "video/mp2t",
	".flv": "video/x-flv", ".3gp": "video/3gpp", ".ogv": "video/ogg",
	".mpg": "video/mpeg", ".mpeg": "video/mpeg", ".wmv": "video/x-ms-wmv",
	".m4a": "audio/mp4", ".mp3": "audio/mpeg", ".flac": "audio/flac",
	".ogg": "audio/ogg", ".opus": "audio/ogg", ".oga": "audio/ogg",
	".wav": "audio/wav", ".aac": "audio/aac", ".ac3": "audio/ac3",
	".mka": "audio/x-matroska", ".wma": "audio/x-ms-wma", ".aiff": "audio/aiff",
}

// MIMEType is the content type of a media file, from its extension.
func MIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if t, ok := mediaTypes[ext]; ok {
		return t
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}
