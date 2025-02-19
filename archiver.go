package arc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

// Maps to handle compression and archival types
var CompressionMap = map[string]archives.Compression{
	"gz":  archives.Gz{},
	"bz2": archives.Bz2{},
	"xz":  archives.Xz{},
	"zst": archives.Zstd{},
	"lz4": archives.Lz4{},
	"br":  archives.Brotli{},
	"lz":  archives.Lzip{},
	"sz":  archives.Sz{},
	"zz":  archives.Zlib{},
}

var ArchivalMap = map[string]archives.Archival{
	"tar": archives.Tar{},
	"zip": archives.Zip{},
}

// check if a path exists
func isExist(path string) bool {
	_, statErr := os.Stat(path)
	return !os.IsNotExist(statErr)
}

// Archive is a function that archives the files in a directory
// dir: the directory to Archive
// outfile: the output file
// compression: the compression to use (gzip, bzip2, etc.)
// archival: the archival to use (tar, zip, etc.)
func Archive(ctx context.Context, dir, outfile string, compression archives.Compression, archival archives.Archival) error {
	logging("Starting the archival process for directory: %s", dir)

	if !isExist(dir) {
		errMsg := fmt.Errorf("directory '%s' does not exist, cannot proceed with archival", dir)
		return errMsg
	}

	// map files on disk to their paths in the archive
	logging("Mapping files in directory: %s", dir)
	files, err := makeDirMap(ctx, dir)
	if err != nil {
		errMsg := fmt.Errorf("error mapping files from directory '%s': %w", dir, err)
		return errMsg
	}
	logging("Successfully mapped files for directory: %s", dir)
	return ArchiveFiles(ctx, files, outfile, compression, archival)
}

func makeDirMap(ctx context.Context, dir string) (files []archives.FileInfo, err error) {
	archiveDirName := filepath.Base(filepath.Clean(dir))
	if dir == "." {
		archiveDirName = ""
	}
	files, err = archives.FilesFromDisk(ctx, nil, map[string]string{
		dir: archiveDirName,
	})
	return
}

func MakeFilesMap(ctx context.Context, files []string, trimDir string) ([]archives.FileInfo, error) {
	mapped := map[string]string{}
	var err error
	trimDir, err = filepath.Abs(trimDir)
	if err != nil {
		return nil, err
	}
	trimDir = trimDir + string(filepath.Separator)
	for _, file := range files {
		file, err = filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		mapped[file] = strings.TrimPrefix(file, trimDir)
	}
	return archives.FilesFromDisk(ctx, nil, mapped)
}

func ArchiveFiles(ctx context.Context, files []archives.FileInfo, outfile string, compression archives.Compression, archival archives.Archival) error {
	// remove outfile
	logging("Removing any existing output file: %s", outfile)
	if err := os.RemoveAll(outfile); err != nil {
		errMsg := fmt.Errorf("failed to remove existing output file '%s': %w", outfile, err)
		return errMsg
	}

	// create the output file we'll write to
	logging("Creating output file: %s", outfile)
	outf, err := os.Create(outfile)
	if err != nil {
		errMsg := fmt.Errorf("error creating output file '%s': %w", outfile, err)
		return errMsg
	}
	defer func() {
		logging("Closing output file: %s", outfile)
		outf.Close()
	}()

	// define the archive format
	logging("Defining the archive format with compression: %T and archival: %T", compression, archival)
	format := archives.CompressedArchive{
		Compression: compression,
		Archival:    archival,
	}

	// create the archive
	logging("Starting archive creation: %s", outfile)
	err = format.Archive(ctx, outf, files)
	if err != nil {
		errMsg := fmt.Errorf("error during archive creation for output file '%s': %w", outfile, err)
		return errMsg
	}
	logging("Archive created successfully: %s", outfile)
	return nil
}
