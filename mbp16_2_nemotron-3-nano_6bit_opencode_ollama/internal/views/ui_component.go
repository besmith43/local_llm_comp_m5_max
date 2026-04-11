package views

// Page holds data for the UI template.
type Page struct {
	VideoFiles []VideoFile
}

// VideoFile represents a video file available for import.
type VideoFile struct {
	Path string // relative path from SOURCE
	Name string // just the filename
	Size int64  // size in bytes
	Ext  string // file extension
}
