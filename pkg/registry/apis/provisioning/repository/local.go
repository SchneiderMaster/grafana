package repository

import (
	"context"
	"path"

	// Git still uses sha1 for the most part: https://git-scm.com/docs/hash-function-transition
	//nolint:gosec
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	provisioning "github.com/grafana/grafana/pkg/apis/provisioning/v0alpha1"
	"github.com/grafana/grafana/pkg/registry/apis/provisioning/safepath"
)

type LocalFolderResolver struct {
	PermittedPrefixes []string
	HomePath          string
}

type InvalidLocalFolderError struct {
	Path           string
	AdditionalInfo string
}

var (
	_ error               = (*InvalidLocalFolderError)(nil)
	_ apierrors.APIStatus = (*InvalidLocalFolderError)(nil)
)

func (e *InvalidLocalFolderError) Error() string {
	return fmt.Sprintf("the path given ('%s') is invalid for a local repository (%s)", e.Path, e.AdditionalInfo)
}

func (e *InvalidLocalFolderError) Status() metav1.Status {
	return metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusBadRequest,
		Reason:  metav1.StatusReasonBadRequest,
		Message: e.Error(),
	}
}

func (r *LocalFolderResolver) LocalPath(p string) (string, error) {
	if len(r.PermittedPrefixes) == 0 {
		return "", &InvalidLocalFolderError{p, "no permitted prefixes were configured"}
	}

	originalPath := p
	if !path.IsAbs(p) {
		var err error
		p, err = safepath.Join(r.HomePath, p)
		if err != nil {
			return "", &InvalidLocalFolderError{originalPath, "the path could not be safely resolved"}
		}
	} else {
		p = safepath.Clean(p)
	}

	for _, permitted := range r.PermittedPrefixes {
		if strings.HasPrefix(p, safepath.Clean(permitted)) {
			return p, nil
		}
	}
	return "", &InvalidLocalFolderError{originalPath, "the path matches no permitted prefix"}
}

var _ Repository = (*localRepository)(nil)

type localRepository struct {
	config   *provisioning.Repository
	resolver *LocalFolderResolver

	// validated path that can be read if not empty
	path string
}

func NewLocal(config *provisioning.Repository, resolver *LocalFolderResolver) *localRepository {
	r := &localRepository{
		config:   config,
		resolver: resolver,
	}
	if config.Spec.Local != nil {
		r.path, _ = resolver.LocalPath(config.Spec.Local.Path)
	}
	return r
}

func (r *localRepository) Config() *provisioning.Repository {
	return r.config
}

// Validate implements provisioning.Repository.
func (r *localRepository) Validate() (fields field.ErrorList) {
	cfg := r.config.Spec.Local
	if cfg == nil {
		fields = append(fields, &field.Error{
			Type:  field.ErrorTypeRequired,
			Field: "spec.local",
		})
		return fields
	}

	// The path value must be set for local provisioning
	if cfg.Path == "" {
		fields = append(fields, field.Required(field.NewPath("spec", "local", "path"),
			"must enter a path to local file"))
	}

	return fields
}

// Test implements provisioning.Repository.
// NOTE: Validate has been called (and passed) before this function should be called
func (r *localRepository) Test(ctx context.Context) (*provisioning.TestResults, error) {
	if r.config.Spec.Local.Path == "" {
		return &provisioning.TestResults{
			Code:    http.StatusBadRequest,
			Success: false,
			Errors: []string{
				"no path is configured",
			},
		}, nil
	}

	_, err := r.resolver.LocalPath(r.config.Spec.Local.Path)
	if err != nil {
		return &provisioning.TestResults{
			Code:    http.StatusBadRequest,
			Success: false,
			Errors: []string{
				err.Error(),
			},
		}, nil
	}

	_, err = os.Stat(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return &provisioning.TestResults{
			Code:    http.StatusBadRequest,
			Success: false,
			Errors: []string{
				fmt.Sprintf("directory not found: %s", r.config.Spec.Local.Path),
			},
		}, nil
	}

	return &provisioning.TestResults{
		Code:    http.StatusOK,
		Success: true,
	}, nil
}

// Test implements provisioning.Repository.
func (r *localRepository) validateRequest(ref string) error {
	if ref != "" {
		return apierrors.NewBadRequest("local repository does not support ref")
	}
	if r.path == "" {
		_, err := r.resolver.LocalPath(r.config.Spec.Local.Path)
		if err != nil {
			return err
		}
		return &apierrors.StatusError{
			ErrStatus: metav1.Status{
				Message: "the service is missing a root path",
				Code:    http.StatusFailedDependency,
			},
		}
	}
	return nil
}

// ReadResource implements provisioning.Repository.
func (r *localRepository) Read(ctx context.Context, filePath string, ref string) (*FileInfo, error) {
	if err := r.validateRequest(ref); err != nil {
		return nil, err
	}

	filePath, err := safepath.Join(r.path, filePath)
	if err != nil {
		return nil, fmt.Errorf("join path: %w", err)
	}

	info, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrFileNotFound
	} else if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	path := strings.TrimPrefix(filePath, safepath.Clean(r.path))
	if info.IsDir() {
		return &FileInfo{
			Path: path,
			Modified: &metav1.Time{
				Time: info.ModTime(),
			},
		}, nil
	}

	//nolint:gosec
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	hash, _, err := r.calculateFileHash(filePath)
	if err != nil {
		return nil, fmt.Errorf("calculate hash of file: %w", err)
	}

	return &FileInfo{
		Path: path,
		Data: data,
		Hash: hash,
		Modified: &metav1.Time{
			Time: info.ModTime(),
		},
	}, nil
}

// ReadResource implements provisioning.Repository.
func (r *localRepository) ReadTree(ctx context.Context, ref string) ([]FileTreeEntry, error) {
	if err := r.validateRequest(ref); err != nil {
		return nil, err
	}

	rootlen := len(r.path)
	entries := make([]FileTreeEntry, 0, 100)
	err := filepath.Walk(r.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		entry := FileTreeEntry{
			Path: strings.TrimLeft(path[rootlen:], "/"),
			Size: info.Size(),
		}
		if !info.IsDir() {
			entry.Blob = true
			entry.Hash, _, err = r.calculateFileHash(path)
			if err != nil {
				return fmt.Errorf("failed to read and calculate hash of path %s: %w", path, err)
			}
		}
		entries = append(entries, entry)
		return err
	})

	return entries, err
}

func (r *localRepository) calculateFileHash(path string) (string, int64, error) {
	// Treats https://securego.io/docs/rules/g304.html
	if !strings.HasPrefix(path, r.path) {
		return "", 0, ErrFileNotFound
	}

	// We've already made sure the path is safe, so we'll ignore the gosec lint.
	//nolint:gosec
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", 0, err
	}

	// TODO: Define what hashing algorithm we want to use for the entire repository. Maybe a config option?
	hasher := sha1.New()
	// TODO: context-aware io.Copy? Is that even possible with a reasonable impl?
	size, err := io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}
	// NOTE: EncodeToString (& hex.Encode for that matter) return lower-case hex.
	return hex.EncodeToString(hasher.Sum(nil)), size, nil
}

func (r *localRepository) Create(ctx context.Context, sanitisedPath string, ref string, data []byte, comment string) error {
	if err := r.validateRequest(ref); err != nil {
		return err
	}

	// before sanitizing, check if the path is a directory
	isDir := strings.HasSuffix(sanitisedPath, "/")
	inputUnsafePath := sanitisedPath // do NOT use in path operations, only for responses!

	sanitisedPath, err := safepath.Join(r.path, sanitisedPath)
	if err != nil {
		return fmt.Errorf("join path: %w", err)
	}

	_, err = os.Stat(sanitisedPath)
	if !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			return apierrors.NewInternalError(fmt.Errorf("failed to check if file exists: %w", err))
		}
		return apierrors.NewAlreadyExists(provisioning.RepositoryResourceInfo.GroupResource(), inputUnsafePath)
	}

	if isDir {
		if data != nil {
			return apierrors.NewBadRequest("data cannot be provided for a directory")
		}

		if err := os.MkdirAll(sanitisedPath, 0700); err != nil {
			return apierrors.NewInternalError(fmt.Errorf("failed to create path: %w", err))
		}

		return nil
	}

	if err := os.MkdirAll(path.Dir(sanitisedPath), 0700); err != nil {
		return apierrors.NewInternalError(fmt.Errorf("failed to create path: %w", err))
	}

	return os.WriteFile(sanitisedPath, data, 0600)
}

func (r *localRepository) Update(ctx context.Context, path string, ref string, data []byte, comment string) error {
	if err := r.validateRequest(ref); err != nil {
		return err
	}

	path, err := safepath.Join(r.path, path)
	if err != nil {
		return fmt.Errorf("join path: %w", err)
	}

	if strings.HasSuffix(path, "/") {
		return apierrors.NewBadRequest("cannot update a directory")
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file does not exist")
	}
	return os.WriteFile(path, data, 0600)
}

func (r *localRepository) Delete(ctx context.Context, path string, ref string, comment string) error {
	if err := r.validateRequest(ref); err != nil {
		return err
	}

	path, err := safepath.Join(r.path, path)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

func (r *localRepository) History(ctx context.Context, path string, ref string) ([]provisioning.HistoryItem, error) {
	return nil, &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Message: "history is not yet implemented",
			Code:    http.StatusNotImplemented,
		},
	}
}

// Webhook implements Repository.
func (r *localRepository) Webhook(ctx context.Context, req *http.Request) (*provisioning.WebhookResponse, error) {
	return &provisioning.WebhookResponse{
		Code: http.StatusAccepted,
		Job: &provisioning.JobSpec{
			Repository: r.Config().GetName(),
			Action:     provisioning.JobActionSync, // sync the latest changes
		},
	}, nil
}
