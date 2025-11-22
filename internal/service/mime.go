package service

import "strings"

func GetContentType(filename string) string {
	// Convert file names to lowercase for case insensitive comparison
	lowerFilename := strings.ToLower(filename)

	switch {
	// Text files
	case strings.HasSuffix(lowerFilename, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".xml"):
		return "application/xml; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".html"), strings.HasSuffix(lowerFilename, ".htm"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".txt"):
		return "text/plain; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".md"):
		return "text/markdown; charset=utf-8"
	case strings.HasSuffix(lowerFilename, ".csv"):
		return "text/csv; charset=utf-8"

	// Image files
	case strings.HasSuffix(lowerFilename, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lowerFilename, ".png"):
		return "image/png"
	case strings.HasSuffix(lowerFilename, ".jpg"), strings.HasSuffix(lowerFilename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lowerFilename, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lowerFilename, ".bmp"):
		return "image/bmp"
	case strings.HasSuffix(lowerFilename, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(lowerFilename, ".svg"), strings.HasSuffix(lowerFilename, ".svgz"):
		return "image/svg+xml"
	case strings.HasSuffix(lowerFilename, ".tiff"), strings.HasSuffix(lowerFilename, ".tif"):
		return "image/tiff"
	case strings.HasSuffix(lowerFilename, ".avif"):
		return "image/avif"

	// Audio files
	case strings.HasSuffix(lowerFilename, ".mp3"):
		return "audio/mpeg"
	case strings.HasSuffix(lowerFilename, ".wav"):
		return "audio/wav"
	case strings.HasSuffix(lowerFilename, ".ogg"):
		return "audio/ogg"
	case strings.HasSuffix(lowerFilename, ".flac"):
		return "audio/flac"
	case strings.HasSuffix(lowerFilename, ".aac"):
		return "audio/aac"
	case strings.HasSuffix(lowerFilename, ".m4a"):
		return "audio/mp4"

	// Video files
	case strings.HasSuffix(lowerFilename, ".mp4"):
		return "video/mp4"
	case strings.HasSuffix(lowerFilename, ".webm"):
		return "video/webm"
	case strings.HasSuffix(lowerFilename, ".ogg"), strings.HasSuffix(lowerFilename, ".ogv"):
		return "video/ogg"
	case strings.HasSuffix(lowerFilename, ".mov"):
		return "video/quicktime"
	case strings.HasSuffix(lowerFilename, ".avi"):
		return "video/x-msvideo"
	case strings.HasSuffix(lowerFilename, ".wmv"):
		return "video/x-ms-wmv"
	case strings.HasSuffix(lowerFilename, ".flv"):
		return "video/x-flv"
	case strings.HasSuffix(lowerFilename, ".mkv"):
		return "video/x-matroska"

	// Font files
	case strings.HasSuffix(lowerFilename, ".woff"):
		return "font/woff"
	case strings.HasSuffix(lowerFilename, ".woff2"):
		return "font/woff2"
	case strings.HasSuffix(lowerFilename, ".ttf"):
		return "font/ttf"
	case strings.HasSuffix(lowerFilename, ".otf"):
		return "font/otf"

	// Archive files
	case strings.HasSuffix(lowerFilename, ".zip"):
		return "application/zip"
	case strings.HasSuffix(lowerFilename, ".rar"):
		return "application/x-rar-compressed"
	case strings.HasSuffix(lowerFilename, ".gz"):
		return "application/gzip"
	case strings.HasSuffix(lowerFilename, ".tar"):
		return "application/x-tar"
	case strings.HasSuffix(lowerFilename, ".7z"):
		return "application/x-7z-compressed"
	case strings.HasSuffix(lowerFilename, ".bz2"):
		return "application/x-bzip2"
	case strings.HasSuffix(lowerFilename, ".xz"):
		return "application/x-xz"

	// Document files
	case strings.HasSuffix(lowerFilename, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lowerFilename, ".doc"):
		return "application/msword"
	case strings.HasSuffix(lowerFilename, ".docx"):
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(lowerFilename, ".xls"):
		return "application/vnd.ms-excel"
	case strings.HasSuffix(lowerFilename, ".xlsx"):
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case strings.HasSuffix(lowerFilename, ".ppt"):
		return "application/vnd.ms-powerpoint"
	case strings.HasSuffix(lowerFilename, ".pptx"):
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case strings.HasSuffix(lowerFilename, ".odt"):
		return "application/vnd.oasis.opendocument.text"
	case strings.HasSuffix(lowerFilename, ".ods"):
		return "application/vnd.oasis.opendocument.spreadsheet"
	case strings.HasSuffix(lowerFilename, ".odp"):
		return "application/vnd.oasis.opendocument.presentation"

	// Other files
	case strings.HasSuffix(lowerFilename, ".rtf"):
		return "application/rtf"
	case strings.HasSuffix(lowerFilename, ".epub"):
		return "application/epub+zip"
	case strings.HasSuffix(lowerFilename, ".apk"):
		return "application/vnd.android.package-archive"
	case strings.HasSuffix(lowerFilename, ".exe"):
		return "application/x-msdownload"
	case strings.HasSuffix(lowerFilename, ".dmg"):
		return "application/x-apple-diskimage"
	case strings.HasSuffix(lowerFilename, ".iso"):
		return "application/x-iso9660-image"

	default:
		return "application/octet-stream"
	}
}
