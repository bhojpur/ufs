package archive

import (
	"os"

	statistics "github.com/bhojpur/host/pkg/statistics"
)

func statDifferent(oldStat *statistics.StatT, newStat *statistics.StatT) bool {
	if !sameFsTime(oldStat.Mtim(), newStat.Mtim()) ||
		oldStat.Mode() != newStat.Mode() ||
		oldStat.Size() != newStat.Size() && !oldStat.Mode().IsDir() {
		return true
	}
	return false
}

func (info *FileInfo) isDir() bool {
	return info.parent == nil || info.stat.Mode().IsDir()
}

func getIno(fi os.FileInfo) (inode uint64) {
	return
}

func hasHardlinks(fi os.FileInfo) bool {
	return false
}
