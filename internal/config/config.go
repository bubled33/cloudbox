package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

type Immutable struct {
	AppName string `koanf:"app_name"`
	HTTP    struct {
		Addr string `koanf:"addr"`
	} `koanf:"http"`
	DB struct {
		DSN string `koanf:"dsn"`
	} `koanf:"db"`
	SMTP struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		Name     string `koanf:"name"`
		Email    string `koanf:"email"`
		Password string `koanf:"password"`
	} `koanf:"smtp"`
	S3 struct {
		Endpoint        string `koanf:"endpoint"`
		Region          string `koanf:"region"`
		Bucket          string `koanf:"bucket"`
		AccessKeyID     string `koanf:"accesskeyid"`
		SecretAccessKey string `koanf:"secretaccesskey"`
		UsePathStyle    bool   `koanf:"use_path_style"`
	} `koanf:"s3"`
	JWT struct {
		SigningKey string `koanf:"signingkey"`
	} `koanf:"jwt"`
}

type Dynamic struct {
	RateLimits struct {
		GlobalRPS int `koanf:"global_rps"`
	} `koanf:"rate_limits"`
	Features struct {
		Previews bool `koanf:"previews"`
	} `koanf:"features"`
}

type Snapshot struct {
	Immutable Immutable
	Dynamic   Dynamic
	Version   int64
}

func Load(basePath, envPath string, httpAddrOverride string) (Snapshot, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(basePath), yaml.Parser()); err != nil {
		fmt.Printf("Warning: failed to load base config: %v\n", err)
	}

	if envPath != "" {
		if err := k.Load(file.Provider(envPath), yaml.Parser()); err != nil {
			fmt.Printf("Warning: failed to load env config: %v\n", err)
		}
	}

	if err := k.Load(env.Provider("CFS_", ".", func(s string) string {
		key := strings.ToLower(strings.TrimPrefix(s, "CFS_"))
		key = strings.ReplaceAll(key, "__", ".")
		fmt.Printf("ENV: %s -> %s\n", s, key)
		return key
	}), nil); err != nil {
		fmt.Printf("Error loading env: %v\n", err)
	}

	if httpAddrOverride != "" {
		raw := []byte("http:\n  addr: " + httpAddrOverride + "\n")
		_ = k.Load(rawbytes.Provider(raw), yaml.Parser())
	}

	fmt.Println("\nAll loaded keys:")
	for _, key := range k.Keys() {
		fmt.Printf("  %s = %v\n", key, k.Get(key))
	}

	var imm Immutable
	var dyn Dynamic
	if err := k.Unmarshal("", &imm); err != nil {
		return Snapshot{}, err
	}
	if err := k.Unmarshal("", &dyn); err != nil {
		return Snapshot{}, err
	}

	fmt.Printf("\nS3 AccessKeyID: '%s'\n", imm.S3.AccessKeyID)
	fmt.Printf("S3 SecretAccessKey: '%s'\n", imm.S3.SecretAccessKey)

	return Snapshot{Immutable: imm, Dynamic: dyn, Version: time.Now().UnixNano()}, nil
}
