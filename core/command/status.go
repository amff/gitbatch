package command

import (
	"errors"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/isacikgoz/gitbatch/core/git"
	log "github.com/sirupsen/logrus"
)

var (
	statusCmdMode string

	statusCommand       = "status"
	statusCmdModeLegacy = "git"
	statusCmdModeNative = "go-git"
)

func shortStatus(r *git.Repository, option string) string {
	args := make([]string, 0)
	args = append(args, statusCommand)
	args = append(args, option)
	args = append(args, "--short")
	out, err := Run(r.AbsPath, "git", args)
	if err != nil {
		log.Warn("Error while status command")
		return "?"
	}
	return out
}

// Status returns the dirty files
func Status(r *git.Repository) ([]*git.File, error) {
	statusCmdMode = statusCmdModeLegacy

	switch statusCmdMode {
	case statusCmdModeLegacy:
		return statusWithGit(r)
	case statusCmdModeNative:
		return statusWithGoGit(r)
	}
	return nil, errors.New("Unhandled status operation")
}

// PlainStatus returns the palin status
func PlainStatus(r *git.Repository) (string, error) {
	args := make([]string, 0)
	args = append(args, "status")
	output, err := Run(r.AbsPath, "git", args)
	if err != nil {
		log.Warn(err)
	}
	re := regexp.MustCompile(`\n?\r`)
	output = re.ReplaceAllString(output, "\n")
	return output, err
}

// LoadFiles function simply commands a git status and collects output in a
// structured way
func statusWithGit(r *git.Repository) ([]*git.File, error) {
	files := make([]*git.File, 0)
	output := shortStatus(r, "--untracked-files=all")
	if len(output) == 0 {
		return files, nil
	}
	fileslist := strings.Split(output, "\n")
	for _, file := range fileslist {
		x := byte(file[0])
		y := byte(file[1])
		relativePathRegex := regexp.MustCompile(`[(\w|/|.|\-)]+`)
		path := relativePathRegex.FindString(file[2:])

		files = append(files, &git.File{
			Name:    path,
			AbsPath: r.AbsPath + string(os.PathSeparator) + path,
			X:       git.FileStatus(x),
			Y:       git.FileStatus(y),
		})
	}
	sort.Sort(git.FilesAlphabetical(files))
	return files, nil
}

func statusWithGoGit(r *git.Repository) ([]*git.File, error) {
	files := make([]*git.File, 0)
	w, err := r.Repo.Worktree()
	if err != nil {
		return files, err
	}
	s, err := w.Status()
	if err != nil {
		return files, err
	}
	for k, v := range s {
		files = append(files, &git.File{
			Name:    k,
			AbsPath: r.AbsPath + string(os.PathSeparator) + k,
			X:       git.FileStatus(v.Staging),
			Y:       git.FileStatus(v.Worktree),
		})
	}
	sort.Sort(git.FilesAlphabetical(files))
	return files, nil
}
