package bean

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const BeansDir = ".beans"

var (
	ErrNotFound     = errors.New("bean not found")
	ErrAmbiguousID  = errors.New("ambiguous ID prefix matches multiple beans")
	ErrNoBeansDir   = errors.New(".beans directory not found")
)

// Store manages beans on the filesystem.
type Store struct {
	Root string // absolute path to .beans directory
}

// NewStore creates a store with the given root path.
func NewStore(root string) *Store {
	return &Store{Root: root}
}

// FindRoot searches upward from the current directory for a .beans directory.
func FindRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		beansPath := filepath.Join(dir, BeansDir)
		if info, err := os.Stat(beansPath); err == nil && info.IsDir() {
			return beansPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", ErrNoBeansDir
		}
		dir = parent
	}
}

// FindAll returns all beans in the store.
func (s *Store) FindAll() ([]*Bean, error) {
	var beans []*Bean

	err := filepath.WalkDir(s.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		bean, err := s.loadBean(path)
		if err != nil {
			return fmt.Errorf("loading %s: %w", path, err)
		}

		beans = append(beans, bean)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return beans, nil
}

// FindByID finds a bean by ID or ID prefix.
func (s *Store) FindByID(idPrefix string) (*Bean, error) {
	beans, err := s.FindAll()
	if err != nil {
		return nil, err
	}

	var matches []*Bean
	for _, b := range beans {
		if strings.HasPrefix(b.ID, idPrefix) {
			matches = append(matches, b)
		}
	}

	switch len(matches) {
	case 0:
		return nil, ErrNotFound
	case 1:
		return matches[0], nil
	default:
		return nil, ErrAmbiguousID
	}
}

// Save writes a bean to disk.
func (s *Store) Save(bean *Bean) error {
	// Set timestamps
	now := time.Now().UTC()
	if bean.CreatedAt == nil {
		bean.CreatedAt = &now
	}
	bean.UpdatedAt = &now

	// Determine the file path
	var path string
	if bean.Path != "" {
		path = filepath.Join(s.Root, bean.Path)
	} else {
		filename := BuildFilename(bean.ID, bean.Slug)
		path = filepath.Join(s.Root, filename)
		bean.Path = filename
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Render and write
	content, err := bean.Render()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// Delete removes a bean from disk.
func (s *Store) Delete(idPrefix string) error {
	bean, err := s.FindByID(idPrefix)
	if err != nil {
		return err
	}

	path := filepath.Join(s.Root, bean.Path)
	return os.Remove(path)
}

// Init creates the .beans directory if it doesn't exist.
func Init(dir string) error {
	beansPath := filepath.Join(dir, BeansDir)
	return os.MkdirAll(beansPath, 0755)
}

// loadBean reads and parses a bean file.
func (s *Store) loadBean(path string) (*Bean, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bean, err := Parse(f)
	if err != nil {
		return nil, err
	}

	// Set metadata from path
	relPath, err := filepath.Rel(s.Root, path)
	if err != nil {
		return nil, err
	}
	bean.Path = relPath

	// Extract ID and slug from filename
	filename := filepath.Base(path)
	bean.ID, bean.Slug = ParseFilename(filename)

	return bean, nil
}

// FullPath returns the absolute path to a bean file.
func (s *Store) FullPath(bean *Bean) string {
	return filepath.Join(s.Root, bean.Path)
}
