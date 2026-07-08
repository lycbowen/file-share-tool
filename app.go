package fileshare

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hicbowen/file-share-tool/frontend"
)

func Run() {
	cfg, err := parseConfig()
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}
	if cfg.ShowVersion {
		fmt.Fprintln(os.Stdout, VersionString())
		return
	}

	share, err := NewShareFS(cfg.TargetDir)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	localIP, err := GetInternalIP()
	if err != nil {
		log.Printf("[WARN] Failed to get local IP: %v", err)
		localIP = "localhost"
	} else {
		log.Println("[INFO] Local IP:", localIP)
	}

	serverURL := buildServerURL(localIP, cfg.Port)
	app := NewAppServer(share, frontend.FS(), serverURL)
	srv := &http.Server{
		Addr:              cfg.Address(),
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("[INFO] Sharing: %s", share.Root())
	log.Printf("[INFO] Server is running at %s", serverURL)
	log.Printf("[INFO] Listening on %s", cfg.Address())
	openBrowser(serverURL, cfg.AutoOpen)
	if err := printAccessQRCode(os.Stdout, serverURL); err != nil {
		log.Printf("[WARN] Failed to print QR code: %v", err)
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		_ = srv.Shutdown(context.Background())
		log.Fatalf("[FATAL] Server error: %v", err)
	}
}
