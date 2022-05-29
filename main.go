package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"

	"github.com/dustin/go-humanize"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/alecthomas/kingpin.v2"
)

func listLfsFiles(workingDir string) (map[string]struct{}, error) {
	cmd := exec.Command("git", "ls-files", "-z", "--full-name", ":(attr:filter=lfs)")
	cmd.Dir = workingDir
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git ls-files: %s", err.Error())
	}

	paths := bytes.Split(out, []byte{'\000'})
	ret := make(map[string]struct{})
	for _, p := range paths {
		path := string(p)
		if path == "" {
			continue
		}

		ret[path] = struct{}{}
	}

	return ret, nil
}

func validateLfsPointer(workingDir string, path string) (bool, error) {
	cat := exec.Command("git", "cat-file", "blob", fmt.Sprintf("HEAD:%s", path))
	cat.Dir = workingDir
	cat.Stderr = os.Stderr
	check := exec.Command("git-lfs", "pointer", "check", "--stdin")
	check.Dir = workingDir

	var err error = nil
	if check.Stdin, err = cat.StdoutPipe(); err != nil {
		return false, fmt.Errorf("failed to validate lfs pointer: %s", err.Error())
	}
	if err := check.Start(); err != nil {
		return false, fmt.Errorf("failed to run git-lfs command: %s", err.Error())
	}
	if err := cat.Run(); err != nil {
		return false, fmt.Errorf("failed to run git cat-file command: %s", err.Error())
	}
	check.Wait()

	return check.ProcessState.ExitCode() == 0, nil
}

func validateLfsFiles(workingDir string, lfsFiles map[string]struct{}) (bool, error) {
	success := true
	for file := range lfsFiles {
		valid, err := validateLfsPointer(workingDir, file)
		if err != nil {
			return false, fmt.Errorf("error while processing '%s': %s\n", file, err.Error())
		} else if !valid {
			fmt.Printf("%s: Invalid LFS Pointer\n", file)
			success = false
		}
	}

	return success, nil
}

func checkLargeFiles(repo *git.Repository, lfsFiles map[string]struct{}, threshold int64) (bool, error) {
	head, err := repo.Head()
	if err != nil {
		return false, err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return false, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return false, err
	}
	success := true

	if err := tree.Files().ForEach(
		func(file *object.File) error {
			if file.Size < threshold {
				return nil
			}

			if _, ok := lfsFiles[file.Name]; !ok {
				fmt.Printf("%s: Untracked large file\n", file.Name)
				success = false
			}
			return nil
		},
	); err != nil {
		return false, err
	}

	return success, nil
}

func run() (bool, error) {
	cwd, _ := os.Getwd()
	var (
		sizeLimit = kingpin.Flag(
			"size-limit",
			"If a positive value is given, the program will check whether "+
				"all the files bigger than the given value are tracked by git-lfs.\n"+
				"(Examples: '10kb', '1MB')").Default("0").String()
		workingDir = kingpin.Arg("working-dir", "git working directory").Default(cwd).String()
	)
	kingpin.Parse()

	var err error
	var repo *git.Repository
	var lfsFiles map[string]struct{}
	var sizeThreshold int64
	success := true

	if size, err := humanize.ParseBytes(*sizeLimit); err != nil {
		return false, fmt.Errorf("invalid size-limit: %s", err.Error())
	} else {
		if size > math.MaxInt64 {
			size = math.MaxInt64
		}

		sizeThreshold = int64(size)
	}

	if repo, err = git.PlainOpen(*workingDir); err != nil {
		return false, fmt.Errorf("failed to open repository: %s\n", err.Error())
	}
	if lfsFiles, err = listLfsFiles(*workingDir); err != nil {
		return false, err
	}

	if success, err = validateLfsFiles(*workingDir, lfsFiles); err != nil {
		return false, err
	}
	if sizeThreshold > 0 {
		if ok, err := checkLargeFiles(repo, lfsFiles, sizeThreshold); err != nil {
			return false, err
		} else {
			success = success && ok
		}
	}

	return success, nil
}

func main() {
	success, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
	if !success {
		os.Exit(1)
	}
}
