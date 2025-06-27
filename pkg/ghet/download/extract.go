package download

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/gookit/color"
	"github.com/mholt/archiver/v4"
	"knative.dev/client/pkg/output/logging"
	"knative.dev/client/pkg/output/tui"
)

func (p Plan) extractArchives(ctx context.Context, args Args) error {
	widgets := tui.NewWidgets(ctx)
	index := githubapi.CreateIndex(p.Assets)
	for _, asset := range index.Archives {
		widgets.Printf("📦 Extracting archive: %s", color.Cyan.Sprintf(asset.Name))
		ar := archiveAsset{Asset: asset, plan: &p}
		lctx := logging.EnsureLogger(ctx, logging.Fields{"asset": asset.Name})
		if err := ar.extract(lctx, args); err != nil {
			return err
		}
	}
	return nil
}

type archiveAsset struct {
	githubapi.Asset
	plan *Plan
}

func (aa archiveAsset) open(ctx context.Context) (fs.FS, error) {
	log := logging.LoggerFrom(ctx)
	fp := aa.plan.cachePath(ctx, aa.Asset)
	log.WithFields(logging.Fields{"archive": fp}).Debug("Opening archive")

	fsys, err := archiver.FileSystem(ctx, fp)
	if err != nil {
		return nil, unexpected(err)
	}

	return fsys, nil
}

func (aa archiveAsset) extract(ctx context.Context, args Args) error {
	fsys, err := aa.open(ctx)
	if err != nil {
		return err
	}

	var binaries []compressedBinary
	if binaries, err = findBinaries(ctx, args, fsys); err != nil {
		return err
	}

	if binaries, err = chooseBinaries(ctx, args, binaries); err != nil {
		return err
	}

	var cv *checksumVerifier
	if args.VerifyInArchive {
		if cv, err = aa.plan.newChecksumVerifier(ctx); err != nil {
			return err
		}
	}

	for _, binary := range binaries {
		if err = extractBinary(ctx, args, fsys, binary, cv); err != nil {
			return err
		}
	}

	return nil
}

func extractBinary(
	ctx context.Context, args Args,
	fsys fs.FS, binary compressedBinary,
	cv *checksumVerifier,
) error {
	var (
		ff  fs.File
		fi  fs.FileInfo
		err error
	)
	widgets := tui.NewWidgets(ctx)
	if fi, err = archiver.TopDirStat(fsys, binary.path); err != nil {
		return unexpected(err)
	}
	if ff, err = archiver.TopDirOpen(fsys, binary.path); err != nil {
		return unexpected(err)
	}
	defer ff.Close()

	label := "🎯 " + binary.Name()
	progress := widgets.NewProgress(int(fi.Size()), tui.Message{
		Text: label, PaddingSize: len(label),
	})
	binaryPath := path.Join(args.Destination, args.ToString())
	if args.MultipleBinaries {
		binaryPath = path.Join(args.Destination, binary.Name())
	}
	hp := hashPair{}
	if err = extractToBinaryPath(binaryPath, args, cv, binary, progress, ff, &hp); err != nil {
		return err
	}

	if hp.actual != nil {
		actualHash := hex.EncodeToString(hp.actual.Sum(nil))
		if hp.expect != actualHash {
			return fmt.Errorf("%w: %s != %s", ErrChecksumMismatch,
				hp.expect, actualHash)
		}
		widgets.Printf("✅ Checksum match the extracted binary")
	}

	if err = os.Chmod(binaryPath, fi.Mode()); err != nil {
		return unexpected(err)
	}

	return nil
}

type hashPair struct {
	actual hash.Hash
	expect string
}

func extractToBinaryPath(
	binaryPath string, args Args,
	cv *checksumVerifier, binary compressedBinary,
	progress tui.Progress, ff fs.File, hp *hashPair,
) error {
	out, err := os.Create(binaryPath)
	if err != nil {
		return unexpected(err)
	}
	var writer io.Writer = out
	if args.VerifyInArchive && cv != nil {
		for _, entry := range cv.entries {
			if entry.Matches(binary.path) {
				hp.actual = entry.newDigest()
				writer = io.MultiWriter(out, hp.actual)
				hp.expect = entry.hash
				break
			}
		}
	}
	if perr := progress.With(func(pc tui.ProgressControl) error {
		_, err = io.Copy(writer, io.TeeReader(ff, pc))
		if err != nil {
			err = unexpected(err)
			pc.Error(err)
			return err
		}
		return nil
	}); perr != nil {
		return perr //nolint:wrapcheck
	}
	if err = out.Close(); err != nil {
		return unexpected(err)
	}
	return nil
}

type compressedBinary struct {
	path string
	fs.FileInfo
}

func (b compressedBinary) String() string {
	return fmt.Sprintf("%s %s %s", b.path, b.Mode().Perm(), b.ModTime())
}

func findBinaries(ctx context.Context, args Args, fsys fs.FS) ([]compressedBinary, error) {
	l := logging.LoggerFrom(ctx)
	binaries := make([]compressedBinary, 0, 1)
	if err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		l.WithFields(logging.Fields{"type": d.Type().Perm()}).
			Debugf("Checking in-archive file: %s", p)
		filename := d.Name()
		var fi fs.FileInfo
		if fi, err = d.Info(); err != nil {
			return unexpected(err)
		}
		if !d.IsDir() && isExecutable(fi.Mode().Perm()) && strings.Contains(filename, args.BaseName) {
			binaries = append(binaries, compressedBinary{p, fi})
		}
		return nil
	}); err != nil {
		return nil, unexpected(err)
	}
	l.WithFields(logging.Fields{"binaries": fmt.Sprintf("%q", binaries)}).
		Debugf("Found %d binaries", len(binaries))
	return binaries, nil
}

func chooseBinaries(ctx context.Context, args Args, binaries []compressedBinary) ([]compressedBinary, error) {
	if args.MultipleBinaries {
		return binaries, nil
	}
	var (
		err    error
		binary compressedBinary
	)
	if binary, err = chooseBinary(ctx, binaries); err != nil {
		return nil, err
	}
	return []compressedBinary{binary}, nil
}

func chooseBinary(ctx context.Context, binaries []compressedBinary) (compressedBinary, error) {
	l := logging.LoggerFrom(ctx)
	if len(binaries) != 1 {
		l.Warnf("Can't choose binary automatically: %q", binaries)
		widgets, err := tui.NewInteractiveWidgets(ctx)
		if err != nil {
			return compressedBinary{}, fmt.Errorf("%w: can't choose binary: %q",
				err, binaries)
		}
		chooser := tui.NewChooser[compressedBinary](widgets)
		return chooser.Choose(binaries, "Choose the binary"), nil
	}
	return binaries[0], nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0o111 != 0
}
