package grouping

import (
	"github.com/jagadeesh-2006/gommit/internals/git"
)

func GetStagedFileNames() ([]string, error) {
	return git.GetStagedFilesNameOnly()
}

func GetShortStat() (string, error) {
	return git.GetStagedShortStat()
}

// missing entries as LineStats{0, 0}.
func GetNumStatBatch(filePaths []string) (map[string]LineStats, error) {
	numStatMap, err := git.GetNumStatBatch(filePaths)
	if err != nil {
		return nil, err
	}

	// Convert git.LineCountPair to grouping.LineStats
	result := make(map[string]LineStats, len(numStatMap))
	for path, pair := range numStatMap {
		result[path] = LineStats{
			Added:   pair.Added,
			Removed: pair.Removed,
		}
	}
	return result, nil
}

func GetStatForGroup(filePaths []string) (string, error) {
	return git.GetDiffStat(filePaths)
}

func GetFileDiffU0(filePath string) (string, error) {
	return git.GetFileDiffU0(filePath)
}

func GetDiffU0(filePaths []string) (string, error) {
	return git.GetDiffU0(filePaths)
}

func GetWordDiff(filePaths []string) (string, error) {
	return git.GetWordDiff(filePaths)
}

func GetDiffForGroup(group FileGroup, filePaths []string) (string, error) {
	switch group {
	case GroupConfig:
		return git.GetWordDiff(filePaths)
	case GroupSkip:
		return "", nil
	default:
		return git.GetDiffU0(filePaths)
	}
}
