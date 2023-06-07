package download

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	"github.com/cardil/ghet/pkg/output/tui"
	slog "github.com/go-eden/slf4go"
	"github.com/gookit/color"
)

func extract(ctx context.Context, assets []githubapi.Asset, args Args) error {
	widgets := tui.WidgetsFrom(ctx)
	index := githubapi.CreateIndex(assets)
	for _, asset := range index.Archives {
		widgets.Printf(ctx, "📦 Extracting archive: %s", color.Cyan.Sprintf(asset.Name))
		ar := archiveAsset{asset}
		ctx = output.EnsureLogger(ctx, slog.Fields{"asset": asset.Name, "type": ar.ty()})
		if err := extractArchive(ctx, ar, args); err != nil {
			return err
		}
	}
	return nil
}

func extractArchive(ctx context.Context, ar archiveAsset, args Args) error {
	r, err := ar.open(ctx, args)
	if err != nil {
		return err
	}
	defer r.Close()

}

type archiveAsset struct {
	githubapi.Asset
}

type archiveType int

const (
	archiveTypeZip archiveType = iota
	archiveTypeTar
	archiveTypeTarGz
	archiveTypeTarBz2
	archiveTypeTarXz
)

func (a archiveAsset) ty() archiveType {
	if strings.HasSuffix(a.Name, ".zip") {
		return archiveTypeZip
	}
	if strings.HasSuffix(a.Name, ".tar") {
		return archiveTypeTar
	}
	if strings.HasSuffix(a.Name, ".tar.gz") || strings.HasSuffix(a.Name, ".tgz") {
		return archiveTypeTarGz
	}
	if strings.HasSuffix(a.Name, ".tar.bz2") || strings.HasSuffix(a.Name, ".tbz2") || strings.HasSuffix(a.Name, ".tbz") {
		return archiveTypeTarBz2
	}
	if strings.HasSuffix(a.Name, ".tar.xz") || strings.HasSuffix(a.Name, ".txz") {
		return archiveTypeTarXz
	}
	return -1
}

func (a archiveAsset) open(ctx context.Context, args Args) (io.ReadCloser, error) {
	log := output.LoggerFrom(ctx)
	log.Debug("Opening archive")
	fp := path.Join(args.Destination, a.Name)
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnexpected, err)
	}

	switch a.ty() {
	case archiveTypeZip:
		zip.OpenReader()
	}
}
