package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Version represents an API version
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns the version string
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// VersionManager handles API versioning
type VersionManager struct {
	versions map[string]Version
	latest   Version
}

// NewVersionManager creates a new version manager
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions: make(map[string]Version),
	}
}

// AddVersion adds a new API version
func (m *VersionManager) AddVersion(version string) error {
	v, err := parseVersion(version)
	if err != nil {
		return err
	}

	m.versions[version] = v
	if v.Major > m.latest.Major ||
		(v.Major == m.latest.Major && v.Minor > m.latest.Minor) ||
		(v.Major == m.latest.Major && v.Minor == m.latest.Minor && v.Patch > m.latest.Patch) {
		m.latest = v
	}

	return nil
}

// GetVersion returns a version by string
func (m *VersionManager) GetVersion(version string) (Version, bool) {
	v, ok := m.versions[version]
	return v, ok
}

// GetLatestVersion returns the latest version
func (m *VersionManager) GetLatestVersion() Version {
	return m.latest
}

// VersionMiddleware handles version routing
func (m *VersionManager) VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract version from path
		path := c.Request.URL.Path
		parts := strings.Split(path, "/")

		if len(parts) < 3 || parts[1] != "api" || !strings.HasPrefix(parts[2], "v") {
			// No version specified, use latest
			version := m.GetLatestVersion().String()
			c.Set("api_version", version)
			c.Next()
			return
		}

		version := strings.TrimPrefix(parts[2], "v")
		if _, ok := m.GetVersion(version); !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Unsupported API version: %s", version),
			})
			c.Abort()
			return
		}

		c.Set("api_version", version)
		c.Next()
	}
}

// VersionGroup creates a versioned route group
func (m *VersionManager) VersionGroup(r *gin.Engine, version string) *gin.RouterGroup {
	v, ok := m.GetVersion(version)
	if !ok {
		panic(fmt.Sprintf("Unsupported API version: %s", version))
	}

	path := fmt.Sprintf("/api/v%s", v.String())
	return r.Group(path)
}

// parseVersion parses a version string into a Version struct
func parseVersion(version string) (Version, error) {
	var v Version
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return v, fmt.Errorf("invalid version format: %s", version)
	}

	var err error
	v.Major, err = parseInt(parts[0])
	if err != nil {
		return v, err
	}

	v.Minor, err = parseInt(parts[1])
	if err != nil {
		return v, err
	}

	v.Patch, err = parseInt(parts[2])
	if err != nil {
		return v, err
	}

	return v, nil
}

// parseInt parses a string to an integer
func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
