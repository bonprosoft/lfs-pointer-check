# lfs-pointer-check

A program to check whether all the files that are supposed to be tracked by `git-lfs` have valid git-lfs pointers.
This program might be effective in collaborative software development, where some people might forget to run `git lfs install`.

It also checks if all the large files in the `HEAD` are tracked by `git-lfs` when the `--size-limit` option is given.

## Install

Use `go install` or download the latest build from the [release page](https://github.com/bonprosoft/lfs-pointer-check/releases).

```sh
# via go install
$ go install github.com/bonprosoft/lfs-pointer-check@latest

# via release
$ wget -c https://github.com/bonprosoft/lfs-pointer-check/releases/download/0.0.1/lfs_pointer_check_linux_amd64.tar.gz -O - | tar -xzv
```

## Usage

```
usage: lfs-pointer-check [<flags>] [<working-dir>]

Flags:
  --help            Show context-sensitive help (also try --help-long and --help-man).
  --size-limit="0"  If a positive value is given, the program will check
                    whether all the files bigger than the given value
                    are tracked by git-lfs. (Examples: '10kb', '1MB')

Args:
  [<working-dir>]  git working directory
```


## Examples

```
$ lfs-pointer-check
images/data.png: Invalid LFS Pointer
images/data2.png: Invalid LFS Pointer

$ lfs-pointer-check --size-limit 1MB
images/data.png: Invalid LFS Pointer
images/data2.png: Invalid LFS Pointer
assets/movie.mp4: Untracked large file
assets/image.gif: Untracked large file

$ lfs-pointer-check --size-limit 1MB ../other_repo/
images/photo.png: Invalid LFS Pointer
```
