package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alpe/ollama-utils/ollama/server"
	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/types/model"
)

var Version string

// ollama environment variables:
// envVars["OLLAMA_DEBUG"],
// envVars["OLLAMA_HOST"],
// envVars["OLLAMA_KEEP_ALIVE"],
// envVars["OLLAMA_MAX_LOADED_MODELS"],
// envVars["OLLAMA_MAX_QUEUE"],
// envVars["OLLAMA_MODELS"],
// envVars["OLLAMA_NUM_PARALLEL"],
// envVars["OLLAMA_NOPRUNE"],
// envVars["OLLAMA_ORIGINS"],
// envVars["OLLAMA_SCHED_SPREAD"],
// envVars["OLLAMA_TMPDIR"],
// envVars["OLLAMA_FLASH_ATTENTION"],
// envVars["OLLAMA_KV_CACHE_TYPE"],
// envVars["OLLAMA_LLM_LIBRARY"],
// envVars["OLLAMA_GPU_OVERHEAD"],
// envVars["OLLAMA_LOAD_TIMEOUT"],
const utilSSHKey = "OLLAMA_UTILS_SSH_KEY"

func main() {
	path, err := defaultKeyPath()
	if err != nil {
		fmt.Printf("failed to get default key path: %s\n", err)
		os.Exit(1)
	}
	var privKeyPath string
	flag.StringVar(&privKeyPath, "ssh-key", path, "Full qualified path to the ollama ssh key for authentication.")
	versionFlag := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()
	args := flag.Args()
	fmt.Printf("started: %v\n", args)

	if *versionFlag {
		fmt.Printf("Version: %s\n", Version)
		return
	}

	if len(args) != 1 {
		fmt.Println("model required")
		os.Exit(1)
	}
	name := model.ParseName(args[0])
	if !name.IsValid() {
		fmt.Println("invalid model name")
		os.Exit(1)
	}

	// persist ssh key from env when given
	if memSSHKey := os.Getenv(utilSSHKey); memSSHKey != "" {
		if _, err = os.Stat(privKeyPath); !errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("key exists in location: %s\n", privKeyPath)
			os.Exit(1)
		}
		if err := os.MkdirAll(filepath.Dir(privKeyPath), 0o700); err != nil {
			fmt.Printf("failed create dirs: %s\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(privKeyPath, []byte(memSSHKey), 0o600); err != nil {
			fmt.Printf("failed to write key: %s\n", err)
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = server.WithPrivKeyPath(ctx, privKeyPath)
	regOpts := server.RegistryOptions{}
	err = server.PullModel(ctx, name.Model, &regOpts, func(rsp api.ProgressResponse) {
		print(".")
	})
	cancel()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := server.PruneLayers(); err != nil {
		fmt.Println(err)
		os.Exit(1)

	}
	fmt.Println("done")
}

const defaultPrivateKey = "id_ed25519"

func defaultKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ollama", defaultPrivateKey), nil
}
