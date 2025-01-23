package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/auth"
)

type registryChallenge struct {
	Realm   string
	Service string
	Scope   string
}

func (r registryChallenge) URL() (*url.URL, error) {
	redirectURL, err := url.Parse(r.Realm)
	if err != nil {
		return nil, err
	}

	values := redirectURL.Query()
	values.Add("service", r.Service)
	for _, s := range strings.Split(r.Scope, " ") {
		values.Add("scope", s)
	}

	values.Add("ts", strconv.FormatInt(time.Now().Unix(), 10))

	nonce, err := auth.NewNonce(rand.Reader, 16)
	if err != nil {
		return nil, err
	}

	values.Add("nonce", nonce)

	redirectURL.RawQuery = values.Encode()
	return redirectURL, nil
}

func getAuthorizationToken(ctx context.Context, challenge registryChallenge) (string, error) {
	redirectURL, err := challenge.URL()
	if err != nil {
		return "", err
	}

	sha256sum := sha256.Sum256(nil)
	data := []byte(fmt.Sprintf("%s,%s,%s", http.MethodGet, redirectURL.String(), base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(sha256sum[:])))))

	headers := make(http.Header)
	signature, err := Sign(ctx, data)
	if err != nil {
		return "", err
	}

	headers.Add("Authorization", signature)

	response, err := makeRequest(ctx, http.MethodGet, redirectURL, headers, nil, &RegistryOptions{})
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("%d: %v", response.StatusCode, err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		if len(body) > 0 {
			return "", fmt.Errorf("%d: %s", response.StatusCode, body)
		} else {
			return "", fmt.Errorf("%d", response.StatusCode)
		}
	}

	var token api.TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return "", err
	}

	return token.Token, nil
}

type privKeyPathContextKey struct{}

func WithPrivKeyPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, privKeyPathContextKey{}, path)
}

func Sign(ctx context.Context, bts []byte) (string, error) {
	keyPath, ok := ctx.Value(privKeyPathContextKey{}).(string)
	if !ok || keyPath == "" {
		return "", errors.New("no ssh key path found")
	}

	privateKeyFile, err := os.ReadFile(keyPath)
	if err != nil {
		slog.Info(fmt.Sprintf("Failed to load private key: %v", err))
		return "", err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyFile)
	if err != nil {
		return "", err
	}

	// get the pubkey, but remove the type
	publicKey := ssh.MarshalAuthorizedKey(privateKey.PublicKey())
	parts := bytes.Split(publicKey, []byte(" "))
	if len(parts) < 2 {
		return "", errors.New("malformed public key")
	}

	signedData, err := privateKey.Sign(rand.Reader, bts)
	if err != nil {
		return "", err
	}

	// signature is <pubkey>:<signature>
	return fmt.Sprintf("%s:%s", bytes.TrimSpace(parts[1]), base64.StdEncoding.EncodeToString(signedData.Blob)), nil
}
