package version

// Version is the current main version
// Configured as dev by default
// and will be set during compile time using ldflags
var Version = "dev"

// Build is the git hash of the current build
// Configured as dev by default
// and will be set during compile time using ldflags
var Build = "dev"