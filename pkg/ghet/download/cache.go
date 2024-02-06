package download

import (
	"context"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"path"
	"strconv"

	configdir "github.com/cardil/ghet/pkg/config/dir"
	githubapi "github.com/cardil/ghet/pkg/github/api"
)

func (p Plan) cachePath(ctx context.Context, asset githubapi.Asset) string {
	dir := path.Join(configdir.Cache(ctx), p.transationID())
	if err := os.MkdirAll(dir, executableMode); err != nil {
		log.Fatal(unexpected(err))
	}
	return path.Join(dir, asset.Name)
}

func (p Plan) cleanCache(ctx context.Context) error {
	fp := path.Join(configdir.Cache(ctx), p.transationID())
	if err := os.RemoveAll(fp); err != nil {
		return unexpected(err)
	}
	return nil
}

func (p Plan) transationID() string {
	h := crc32.NewIEEE()
	for _, asset := range p.Assets {
		repr := fmt.Sprintf("%#v", asset)
		_, _ = h.Write([]byte(repr))
	}
	return strconv.Itoa(int(h.Sum32()))
}
