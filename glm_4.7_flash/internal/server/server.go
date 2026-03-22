package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"plex-importor/internal/handler"
	"plex-importor/internal/mover"
	"plex-importor/internal/scanner"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

// SourceScanner is the main scanner type
type SourceScanner struct {
	sourcePath string
}

// NewSourceScanner creates a new SourceScanner instance
func NewSourceScanner(sourcePath string) *SourceScanner {
	return &SourceScanner{
		sourcePath: sourcePath,
	}
}

// ScanFolders scans the source directory recursively and returns folders containing video files
func (s *SourceScanner) ScanFolders() ([]scanner.FolderInfo, error) {
	return scanner.ScanFolders(s.sourcePath)
}

type Server struct {
	port       int
	sourcePath string
	destinationPath string
	scanner    *SourceScanner
	mover      *mover.Mover
	handler    *handler.MetadataHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	sourcePath := os.Getenv("SOURCE")
	destinationPath := os.Getenv("DESTINATION")

	// Initialize components
	scanner := NewSourceScanner(sourcePath)
	mover := mover.NewMover(destinationPath)
	handler := handler.NewMetadataHandler(mover)

	server := &Server{
		port:            port,
		sourcePath:      sourcePath,
		destinationPath: destinationPath,
		scanner:         scanner,
		mover:           mover,
		handler:         handler,
	}

	// Declare Server config
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Set context values for handlers
	httpServer.Handler = server.setContextValues(httpServer.Handler)

	return httpServer
}

// setContextValues sets the context values for handlers
func (s *Server) setContextValues(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, "sourcePath", s.sourcePath)
		ctx = context.WithValue(ctx, "destinationPath", s.destinationPath)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
