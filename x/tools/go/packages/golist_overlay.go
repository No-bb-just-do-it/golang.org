package packages

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

// processGolistOverlay provides rudimentary support for adding
// files that don't exist on disk to an overlay. The results can be
// sometimes incorrect.
// TODO(matloob): Handle unsupported cases, including the following:
// - test files
// - adding test and non-test files to test variants of packages
// - determining the correct package to add given a new import path
// - creating packages that don't exist
func processGolistOverlay(cfg *Config, response *driverResponse) (modifiedPkgs, needPkgs []string, err error) {
	havePkgs := make(map[string]string) // importPath -> non-test package ID
	needPkgsSet := make(map[string]bool)
	modifiedPkgsSet := make(map[string]bool)

	for _, pkg := range response.Packages {
		// This is an approximation of import path to id. This can be
		// wrong for tests, vendored packages, and a number of other cases.
		havePkgs[pkg.PkgPath] = pkg.ID
	}

outer:
	for path, contents := range cfg.Overlay {
		base := filepath.Base(path)
		if strings.HasSuffix(path, "_test.go") {
			// Overlays don't support adding new test files yet.
			// TODO(matloob): support adding new test files.
			continue
		}
		dir := filepath.Dir(path)
		for _, pkg := range response.Packages {
			var dirContains, fileExists bool
			for _, f := range pkg.GoFiles {
				if sameFile(filepath.Dir(f), dir) {
					dirContains = true
				}
				if filepath.Base(f) == base {
					fileExists = true
				}
			}
<<<<<<< HEAD
			if dirContains {
=======
			// The overlay could have included an entirely new package.
			isNewPackage := extractPackage(pkg, path, contents)
			if dirContains || isNewPackage {
>>>>>>> bd25a1f6d07d2d464980e6a8576c1ed59bb3950a
				if !fileExists {
					pkg.GoFiles = append(pkg.GoFiles, path) // TODO(matloob): should the file just be added to GoFiles?
					pkg.CompiledGoFiles = append(pkg.CompiledGoFiles, path)
					modifiedPkgsSet[pkg.ID] = true
				}
				imports, err := extractImports(path, contents)
				if err != nil {
					// Let the parser or type checker report errors later.
					continue outer
				}
				for _, imp := range imports {
					_, found := pkg.Imports[imp]
					if !found {
						needPkgsSet[imp] = true
						// TODO(matloob): Handle cases when the following block isn't correct.
						// These include imports of test variants, imports of vendored packages, etc.
						id, ok := havePkgs[imp]
						if !ok {
							id = imp
						}
						pkg.Imports[imp] = &Package{ID: id}
					}
				}
				continue outer
			}
		}
	}

	needPkgs = make([]string, 0, len(needPkgsSet))
	for pkg := range needPkgsSet {
		needPkgs = append(needPkgs, pkg)
	}
	modifiedPkgs = make([]string, 0, len(modifiedPkgsSet))
	for pkg := range modifiedPkgsSet {
		modifiedPkgs = append(modifiedPkgs, pkg)
	}
	return modifiedPkgs, needPkgs, err
}

func extractImports(filename string, contents []byte) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, contents, parser.ImportsOnly) // TODO(matloob): reuse fileset?
	if err != nil {
		return nil, err
	}
	var res []string
	for _, imp := range f.Imports {
		quotedPath := imp.Path.Value
		path, err := strconv.Unquote(quotedPath)
		if err != nil {
			return nil, err
		}
		res = append(res, path)
	}
	return res, nil
}
<<<<<<< HEAD
=======

// extractPackage attempts to extract a package defined in an overlay.
//
// If the package has errors and has no Name, GoFiles, or Imports,
// then it's possible that it doesn't yet exist on disk.
func extractPackage(pkg *Package, filename string, contents []byte) bool {
	// TODO(rstambler): Check the message of the actual error?
	// It differs between $GOPATH and module mode.
	if len(pkg.Errors) != 1 {
		return false
	}
	if pkg.Name != "" || pkg.ExportFile != "" {
		return false
	}
	if len(pkg.GoFiles) > 0 || len(pkg.CompiledGoFiles) > 0 || len(pkg.OtherFiles) > 0 {
		return false
	}
	if len(pkg.Imports) > 0 {
		return false
	}
	f, err := parser.ParseFile(token.NewFileSet(), filename, contents, parser.PackageClauseOnly) // TODO(matloob): reuse fileset?
	if err != nil {
		return false
	}
	// TODO(rstambler): This doesn't work for main packages.
	if filepath.Base(pkg.PkgPath) != f.Name.Name {
		return false
	}
	pkg.Name = f.Name.Name
	pkg.Errors = nil
	return true
}
>>>>>>> bd25a1f6d07d2d464980e6a8576c1ed59bb3950a
