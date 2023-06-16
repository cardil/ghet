package download

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/1set/gut/yos"
	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	slog "github.com/go-eden/slf4go"
)

func (p Plan) moveBinaries(ctx context.Context, args Args) error {
	l := output.LoggerFrom(ctx)
	index := githubapi.CreateIndex(p.Assets)
	binaryName := args.FileName.ToString()
	for _, binary := range index.Binaries {
		if len(index.Binaries) > 1 {
			binaryName = binary.Name
		}
		l.WithFields(slog.Fields{"binary": binary}).Debug("Moving binary")
		source := p.cachePath(ctx, binary)
		target := path.Join(args.Destination, binaryName)
		if err := yos.MoveFile(source, target); err != nil {
			return unexpected(err)
		}

		if strings.Contains(binary.ContentType, "octet-stream") {
			if err := os.Chmod(target, executableMode); err != nil {
				return unexpected(err)
			}
		}
	}
	return nil
}
