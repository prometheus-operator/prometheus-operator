// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

type GitPackage struct {
	Source *deps.Git
}

func NewGitPackage(source *deps.Git) Interface {
	return &GitPackage{
		Source: source,
	}
}

func downloadGitHubArchive(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	color.Cyan("GET %s %d", url, resp.StatusCode)
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func gzipUntar(dst string, r io.Reader, subDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// strip the two first components of the path
		parts := strings.SplitAfterN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		suffix := parts[1]
		prefix := dst

		// reconstruct the target parh for the archive entry
		target := filepath.Join(prefix, suffix)

		// if subdir is provided and target is not under it, skip it
		subDirPath := filepath.Join(prefix, subDir)
		if subDir != "" && !strings.HasPrefix(target, subDirPath) {
			continue
		}

		// check the file type
		switch header.Typeflag {

		// create directories as needed
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// Explicitly release the file handle inside the inner loop
			// Using defer would accumulate an unbounded quantity of
			// handles and release them all at once at function end.
			f.Close()

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), os.FileMode(header.Mode)); err != nil {
				return err
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
}

func remoteResolveRef(ctx context.Context, remote string, ref string) (string, error) {
	b := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--heads", "--tags", "--refs", "--quiet", remote, ref)
	cmd.Stdin = os.Stdin
	cmd.Stdout = b
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	commitShaPattern := regexp.MustCompile("^([0-9a-f]{40,})\\b")
	commitSha := commitShaPattern.FindString(b.String())
	return commitSha, nil
}

func (p *GitPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	tmpDir, err := ioutil.TempDir(filepath.Join(dir, ".tmp"), fmt.Sprintf("jsonnetpkg-%s-%s", strings.Replace(name, "/", "-", -1), version))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tmp dir")
	}
	defer os.RemoveAll(tmpDir)

	// Optimization for GitHub sources: download a tarball archive of the requested
	// version instead of cloning the entire repository.
	isGitHubRemote, err := regexp.MatchString(`^(https|ssh)://github\.com/.+$`, p.Source.Remote())
	if isGitHubRemote {
		// Let git ls-remote decide if "version" is a ref or a commit SHA in the unlikely
		// but possible event that a ref is comprised of 40 or more hex characters
		commitSha, err := remoteResolveRef(ctx, p.Source.Remote(), version)

		// If the ref resolution failed and "version" looks like a SHA,
		// assume it is one and proceed.
		commitShaPattern := regexp.MustCompile("^([0-9a-f]{40,})$")
		if commitSha == "" && commitShaPattern.MatchString(version) {
			commitSha = version
		}

		archiveUrl := fmt.Sprintf("%s/archive/%s.tar.gz", p.Source.Remote(), commitSha)
		archiveFilepath := fmt.Sprintf("%s.tar.gz", tmpDir)

		defer os.Remove(archiveFilepath)
		err = downloadGitHubArchive(archiveFilepath, archiveUrl)
		if err == nil {
			r, err := os.Open(archiveFilepath)
			defer r.Close()
			if err == nil {
				// Extract the sub-directory (if any) from the archive
				// If none specified, the entire archive is unpacked
				err = gzipUntar(tmpDir, r, p.Source.Subdir)

				// Move the extracted directory to its final destination
				if err == nil {
					if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
						panic(err)
					}
					if err := os.Rename(path.Join(tmpDir, p.Source.Subdir), destPath); err != nil {
						panic(err)
					}
				}
			}
		}

		if err == nil {
			return commitSha, nil
		}

		// The repository may be private or the archive download may not work
		// for other reasons. In any case, fall back to the slower git-based installation.
		color.Yellow("archive install failed: %s", err)
		color.Yellow("retrying with git...")
	}

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", p.Source.Remote())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Attempt shallow fetch at specific revision
	cmd = exec.CommandContext(ctx, "git", "fetch", "--tags", "--depth", "1", "origin", version)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		// Fall back to normal fetch (all revisions)
		cmd = exec.CommandContext(ctx, "git", "fetch", "origin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = tmpDir
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	// Sparse checkout optimization: if a Subdir is specified,
	// there is no need to do a full checkout
	if p.Source.Subdir != "" {
		cmd = exec.CommandContext(ctx, "git", "config", "core.sparsecheckout", "true")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = tmpDir
		err = cmd.Run()
		if err != nil {
			return "", err
		}

		glob := []byte(p.Source.Subdir + "/*\n")
		err = ioutil.WriteFile(filepath.Join(tmpDir, ".git", "info", "sparse-checkout"), glob, 0644)
		if err != nil {
			return "", err
		}
	}

	cmd = exec.CommandContext(ctx, "git", "-c", "advice.detachedHead=false", "checkout", version)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer(nil)
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Stdout = b
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	commitHash := strings.TrimSpace(b.String())

	err = os.RemoveAll(path.Join(tmpDir, ".git"))
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(path.Dir(destPath), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "failed to create parent path")
	}

	err = os.RemoveAll(destPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to clean previous destination path")
	}

	err = os.Rename(path.Join(tmpDir, p.Source.Subdir), destPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to move package")
	}

	return commitHash, nil
}
