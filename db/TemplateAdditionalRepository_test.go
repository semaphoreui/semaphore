package db

import (
	"testing"
)

func TestTemplateAdditionalRepository_Validate_ValidPath(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "libs/mylib",
	}

	err := tar.Validate()
	if err != nil {
		t.Fatalf("expected no error for valid path, got: %v", err)
	}
}

func TestTemplateAdditionalRepository_Validate_EmptyPath(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "",
	}

	err := tar.Validate()
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestTemplateAdditionalRepository_Validate_PathWithLeadingSlash(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "/libs/mylib",
	}

	err := tar.Validate()
	if err != nil {
		t.Fatalf("expected no error, leading slash should be trimmed, got: %v", err)
	}

	if tar.Path != "libs/mylib" {
		t.Fatalf("expected path 'libs/mylib' after trim, got '%s'", tar.Path)
	}
}

func TestTemplateAdditionalRepository_Validate_PathWithTrailingSlash(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "libs/mylib/",
	}

	err := tar.Validate()
	if err != nil {
		t.Fatalf("expected no error, trailing slash should be trimmed, got: %v", err)
	}

	if tar.Path != "libs/mylib" {
		t.Fatalf("expected path 'libs/mylib' after trim, got '%s'", tar.Path)
	}
}

func TestTemplateAdditionalRepository_Validate_PathWithInvalidCharacters(t *testing.T) {
	invalidPaths := []string{
		"libs/my lib",     // space
		"libs/my@lib",     // @
		"libs/my$lib",     // $
		"libs/my&lib",     // &
		"libs/my*lib",     // *
		"libs/my(lib)",    // parentheses
		"libs/my[lib]",    // brackets
		"libs/my{lib}",    // braces
		"libs/my;lib",     // semicolon
		"libs/my:lib",     // colon
		"libs/my'lib",     // quote
		"libs/my\"lib",    // double quote
		"libs/my<lib>",    // angle brackets
		"libs/my|lib",     // pipe
	}

	for _, invalidPath := range invalidPaths {
		tar := &TemplateAdditionalRepository{
			Path: invalidPath,
		}

		err := tar.Validate()
		if err == nil {
			t.Fatalf("expected error for invalid path '%s'", invalidPath)
		}
	}
}

func TestTemplateAdditionalRepository_Validate_PathWithDirectoryTraversal(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "libs/../../../etc/passwd",
	}

	err := tar.Validate()
	if err == nil {
		t.Fatal("expected error for path with directory traversal")
	}
}

func TestTemplateAdditionalRepository_Validate_PathWithBackslash(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "libs\\mylib",
	}

	err := tar.Validate()
	if err == nil {
		t.Fatal("expected error for path with backslash")
	}
}

func TestTemplateAdditionalRepository_Validate_PathTooLong(t *testing.T) {
	longPath := ""
	for i := 0; i < 260; i++ {
		longPath += "a"
	}

	tar := &TemplateAdditionalRepository{
		Path: longPath,
	}

	err := tar.Validate()
	if err == nil {
		t.Fatal("expected error for path exceeding 255 characters")
	}
}

func TestTemplateAdditionalRepository_Validate_ReservedPaths(t *testing.T) {
	reservedPaths := []string{".", "..", "tmp", "cache", "logs", "log"}

	for _, reserved := range reservedPaths {
		tar := &TemplateAdditionalRepository{
			Path: reserved,
		}

		err := tar.Validate()
		if err == nil {
			t.Fatalf("expected error for reserved path '%s'", reserved)
		}
	}
}

func TestTemplateAdditionalRepository_Validate_InvalidRepositoryType(t *testing.T) {
	tar := &TemplateAdditionalRepository{
		Path: "libs/mylib",
		Repository: &Repository{
			GitURL: "file:///path/to/repo",
		},
	}

	err := tar.Validate()
	if err == nil {
		t.Fatal("expected error for non-git repository type")
	}
}

func TestTemplateAdditionalRepository_Validate_ValidRepositoryTypes(t *testing.T) {
	validURLs := map[string]string{
		"git":   "git://github.com/test/repo",
		"ssh":   "git@github.com:test/repo.git",
		"https": "https://github.com/test/repo",
	}

	for repoType, gitURL := range validURLs {
		tar := &TemplateAdditionalRepository{
			Path: "libs/mylib",
			Repository: &Repository{
				GitURL: gitURL,
			},
		}

		err := tar.Validate()
		if err != nil {
			t.Fatalf("expected no error for repository type '%s' with URL '%s', got: %v", repoType, gitURL, err)
		}
	}
}

func TestTemplateAdditionalRepository_Validate_ValidPathVariations(t *testing.T) {
	validPaths := []string{
		"libs/mylib",
		"libs/my-lib",
		"libs/my_lib",
		"libs/my-lib_v2",
		"a/b/c/d/e",
		"mylib",
		"libs/my-lib-123",
		"libs/MY-LIB",
	}

	for _, validPath := range validPaths {
		tar := &TemplateAdditionalRepository{
			Path: validPath,
		}

		err := tar.Validate()
		if err != nil {
			t.Fatalf("expected no error for valid path '%s', got: %v", validPath, err)
		}
	}
}
