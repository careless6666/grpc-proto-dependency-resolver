//go:generate mockgen -source=parser.go -destination=mock/parser-mock.go -package mock

package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

//type IDepsFileParser interface {
//	GetDeps(path string) ([]Dependency, error)
//}

type IFileReader interface {
	ReadFile(filePath string) ([]byte, error)
}

const (
	DependencyTypeGit  = iota
	DependencyTypeURL  = iota
	DependencyTypePath = iota
)

type Dependency struct {
	//locale | remote
	Type int
	Path string
	// for url | path
	DestinationPath string
	Version         *VersionInfo
}

type VersionInfo struct {
	Tag    string
	Commit string
}

type DepsFileParser struct {
	fileReader IFileReader
}

func NewFileParser(fileReader IFileReader) *DepsFileParser {
	return &DepsFileParser{
		fileReader: fileReader,
	}
}

func (f *DepsFileParser) GetDeps(path string) ([]Dependency, error) {
	content, err := f.fileReader.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, errors.New("no dependencies found, empty file")
	}

	depsStr := strings.Split(string(content), "\n")
	//check version
	matchedVersion, err := regexp.Match(`(version)(:)( )(v\d+)`, []byte(depsStr[0]))
	if err != nil {
		return nil, err
	}

	if !matchedVersion {
		return nil, errors.New("invalid dependencies found, not a version")
	}

	if len(depsStr) < 3 {
		return nil, errors.New("invalid dependencies file, rows count less than 3")
	}
	//check deps
	if depsStr[1] != "deps:" {
		return nil, errors.New("invalid dependencies file, \"deps:\" block not found")
	}

	result := make([]Dependency, 0)
	//get deps
	for _, depStr := range depsStr[2:] {
		dep, err := ParseDepsLine(depStr)
		if err != nil {
			return nil, err
		}

		result = append(result, *dep)
	}

	fmt.Println(content)

	return result, nil
}

func ParseDepsLine(dependency string) (*Dependency, error) {
	dependency = strings.TrimSpace(dependency)

	if dependency == "" {
		return nil, errors.New("empty dependency")
	}

	matchedGit, err := regexp.Match(`- git: `, []byte(dependency))
	if err != nil {
		return nil, err
	}

	if matchedGit {
		return getGitDeps(dependency[6:])
	}

	matchedURL, err := regexp.Match(`- url: `, []byte(dependency))
	if err != nil {
		return nil, err
	}

	if matchedURL {
		return getUrlDeps(dependency[6:])
	}

	matchedFile, err := regexp.Match(`- path: `, []byte(dependency))
	if err != nil {
		return nil, err
	}

	if matchedFile {
		return getFileDeps(dependency[7:])
	}

	return nil, nil
}

func getGitDeps(dependency string) (*Dependency, error) {
	depPaths := strings.Split(dependency, " ")
	if len(depPaths) != 2 {
		return nil, errors.New("invalid dependency, have to by pattern \"- git: github.com/repo/file.proto v0.0.0-20211005231101-409e134ffaac\"")
	}

	return nil, nil
}

func getUrlDeps(dependency string) (*Dependency, error) {
	depPaths := strings.Split(dependency, " ")
	if len(depPaths) != 3 {
		return nil, errors.New("invalid dependency, have to by pattern \"- url: https://github.com/repo/file.proto ./github.com/repo/file.proto v1\"")
	}

	matchedProtoFileURL, err := regexp.Match(`(http:\/\/)(.*)(.)(proto)|(https:\/\/)(.*)(.)(proto)`, []byte(depPaths[0]))
	if err != nil {
		return nil, err
	}

	if matchedProtoFileURL {
		return &Dependency{
			Type:            DependencyTypeURL,
			Path:            depPaths[0],
			DestinationPath: depPaths[1],
			Version: &VersionInfo{
				Tag:    depPaths[2],
				Commit: "",
			},
		}, nil
	}

	return nil, errors.New("invalid dependency, expected URL to proto file")
}

func getFileDeps(dependency string) (*Dependency, error) {
	depPaths := strings.Split(dependency, " ")
	if len(depPaths) != 3 {
		return nil, errors.New("invalid dependency, have to by pattern \"- path: /var/github.com/repo/file.proto ./github.com/repo/file.proto v1\"")
	}

	return &Dependency{
		Type:            DependencyTypePath,
		Path:            depPaths[0],
		DestinationPath: depPaths[1],
		Version: &VersionInfo{
			Tag: depPaths[2],
		},
	}, nil
}
