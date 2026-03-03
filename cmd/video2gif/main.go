package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oarkflow/video2gif/internal/config"
	"github.com/oarkflow/video2gif/internal/converter"
	"github.com/oarkflow/video2gif/internal/server"
)

const version = "1.0.0"

const banner = `
 ██╗   ██╗██╗██████╗ ███████╗ ██████╗ ██████╗ ██╗███████╗
 ██║   ██║██║██╔══██╗██╔════╝██╔═══██╗╚════██╗██║██╔════╝
 ██║   ██║██║██║  ██║█████╗  ██║   ██║ █████╔╝██║█████╗
 ╚██╗ ██╔╝██║██║  ██║██╔══╝  ██║   ██║██╔═══╝ ██║██╔══╝
  ╚████╔╝ ██║██████╔╝███████╗╚██████╔╝███████╗██║██║
   ╚═══╝  ╚═╝╚═════╝ ╚══════╝ ╚═════╝ ╚══════╝╚═╝╚═╝
  Production Video→GIF Converter  v` + version + `
`

func main() {
	fmt.Print(banner)

	var (
		cfgPath    = flag.String("config", "config.json", "Path to config.json")
		port       = flag.Int("port", 0, "Override server port")
		cliMode    = flag.Bool("cli", false, "CLI mode: convert one file and exit")
		inputFile  = flag.String("input", "", "Input video (CLI mode)")
		outputFile = flag.String("output", "", "Output GIF (CLI mode)")
		profileName = flag.String("profile", "", "Profile to use (CLI mode)")
	)
	flag.Parse()

	// Load config
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Printf("⚠  Using default config (%v)", err)
		cfg = config.Default()
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}

	// Check ffmpeg
	ffVer, fpVer, err := converter.CheckFFmpeg()
	if err != nil {
		log.Fatalf("❌ %v\n   Install: apt install ffmpeg  |  brew install ffmpeg  |  https://ffmpeg.org/download.html", err)
	}
	log.Printf("✅ %s", ffVer)
	log.Printf("✅ %s", fpVer)

	// CLI mode
	if *cliMode {
		if *inputFile == "" || *outputFile == "" {
			log.Fatal("❌ --input and --output are required in CLI mode")
		}
		pName := *profileName
		if pName == "" {
			pName = cfg.DefaultProfile
		}
		profile, ok := cfg.GetProfile(pName)
		if !ok {
			log.Fatalf("❌ Profile %q not found in config", pName)
		}
		runCLI(cfg, *inputFile, *outputFile, profile)
		return
	}

	// Web server mode
	srv := server.New(cfg, *cfgPath)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Router(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server: http://%s", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ %v", err)
		}
	}()

	<-quit
	log.Println("⏳ Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	srv.Shutdown()
	log.Println("✅ Bye!")
}

func runCLI(cfg *config.Config, input, output string, profile config.GifProfile) {
	conv := converter.NewConverter(cfg)
	job := &converter.ConversionJob{
		ID: "cli", InputPath: input, OutputPath: output,
		Profile: profile, CreatedAt: time.Now(),
	}
	result, err := conv.Convert(context.Background(), job)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	log.Printf("✅ %s  (%s in %s, %d frames)",
		result.OutputPath,
		converter.FormatBytes(result.OutputSize),
		result.Duration.Round(time.Millisecond),
		result.FrameCount,
	)
}
