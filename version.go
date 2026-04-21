package main

// Version is embedded at build time via -ldflags "-X main.Version=vX.Y.Z".
// In local dev builds this stays "dev" and update checks are skipped.
var Version = "dev"
