package download

import (
	"bufio"
	"context"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"emperror.dev/errors"
	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	"github.com/cardil/ghet/pkg/output/tui"
	slog "github.com/go-eden/slf4go"
	"github.com/gookit/color"
)

// ErrTooManyChecksums is returned when there are more than one checksum.
var ErrTooManyChecksums = errors.New("too many checksums")

// ErrNoChecksum is returned when there are no checksums.
var ErrNoChecksum = errors.New("no checksum")

// ErrUnknownChecksumAlgorithm is returned when the checksum algorithm is unknown.
var ErrUnknownChecksumAlgorithm = errors.New("unknown checksum algorithm")

// ErrChecksumMismatch is returned when the checksum does not match.
var ErrChecksumMismatch = errors.New("checksum mismatch")

// ErrNotVerifiedAssets is returned when there are no verified assets.
var ErrNotVerifiedAssets = errors.New("not verified assets")

// ErrInvalidChecksumLine is returned when the checksum line is invalid.
var ErrInvalidChecksumLine = errors.New("invalid checksum line")

var bsdStyleChecksums = regexp.MustCompile(`^(SHA[0-9]{1,3})\s+\(([^)]+)\)\s+=\s+([a-fA-F0-9]{32,128})$`)

func (p Plan) verifyChecksums(ctx context.Context) error {
	widgets := tui.WidgetsFrom(ctx)
	cs, err := p.newChecksumVerifier(ctx)
	if err != nil {
		if errors.Is(err, ErrNoChecksum) {
			widgets.Printf(ctx, "⚠️ No checksums found. Skipping verification")
			return nil
		}
		return err
	}

	index := githubapi.CreateIndex(p.Assets)
	artifacts := make([]githubapi.Asset, 0, len(index.Archives)+len(index.Binaries))
	artifacts = append(append(artifacts, index.Archives...), index.Binaries...)
	err = cs.verify(ctx, artifacts, func(curr githubapi.Asset) string {
		return path.Dir(p.cachePath(ctx, curr))
	})
	if err != nil {
		return err
	}

	widgets.Printf(ctx, "✅ All checksums match the downloaded assets")

	return nil
}

func (p Plan) newChecksumVerifier(ctx context.Context) (*checksumVerifier, error) {
	l := output.LoggerFrom(ctx)
	index := githubapi.CreateIndex(p.Assets)
	if len(index.Checksums) == 0 {
		l.Debug("No checksums to verify")
		return nil, ErrNoChecksum
	}

	ca := index.Checksums[0]

	if len(index.Checksums) > 1 {
		iwidgets, err := tui.Interactive[githubapi.Asset](ctx)
		if err != nil {
			if errors.Is(err, tui.ErrNotInteractive) {
				l.Errorf("Number of checksums is %d. Expected just one.", len(index.Checksums))
				return nil, fmt.Errorf("%w: %d", ErrTooManyChecksums, len(index.Checksums))
			}
			return nil, unexpected(err)
		}
		selected := iwidgets.Choose(ctx, index.Checksums,
			"⚠️ More than one checksum file found. Choose proper one")
		for _, c := range index.Checksums {
			if c == selected {
				ca = c
				break
			}
		}
	}
	artifacts := make([]githubapi.Asset, 0, len(index.Archives)+len(index.Binaries))
	artifacts = append(append(artifacts, index.Archives...), index.Binaries...)
	if len(artifacts) == 0 {
		l.Errorf("No assets to verify")
		return nil, fmt.Errorf("%w: %d", ErrNotVerifiedAssets, len(index.Binaries))
	}

	l = l.WithFields(slog.Fields{"checksum": ca.Name})
	l.Debug("Verifying checksum")

	parser := checksumParser{Asset: ca, plan: &p}
	verifier, err := parser.parse(ctx)
	if err != nil {
		return nil, err
	}
	return verifier, nil
}

type checksumParser struct {
	githubapi.Asset
	plan *Plan
	*checksumVerifier
}

func (p *checksumParser) parse(ctx context.Context) (*checksumVerifier, error) {
	l := output.LoggerFrom(ctx)
	fp := p.plan.cachePath(ctx, p.Asset)
	l.Debugf("Parsing checksum: %s", fp)
	if _, ferr := os.Stat(fp); ferr != nil {
		return nil, unexpected(ferr)
	}
	file, ferr := os.Open(fp)
	if ferr != nil {
		return nil, unexpected(ferr)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	p.checksumVerifier = &checksumVerifier{
		entries: make([]checksumEntry, 0, 1),
	}
	for scanner.Scan() {
		if err := p.parseLine(ctx, scanner.Text()); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, unexpected(err)
	}

	return p.checksumVerifier, nil
}

func (p *checksumParser) parseLine(ctx context.Context, line string) error {
	var entry checksumEntry
	if bsdStyleChecksums.MatchString(line) {
		entry = p.parseBSDStyleChecksum(ctx, line)
	} else {
		e, err := p.parseRegularChecksum(line)
		if err != nil {
			return err
		}
		entry = e
	}
	p.checksumVerifier.entries = append(p.checksumVerifier.entries, entry)
	return nil
}

const (
	minChecksumFields = 1
	maxChecksumFields = 2
)

func (p *checksumParser) parseRegularChecksum(line string) (checksumEntry, error) {
	fields := strings.Fields(line)
	if len(fields) > maxChecksumFields || len(fields) < minChecksumFields {
		return checksumEntry{}, unexpected(fmt.Errorf("%w: %s", ErrInvalidChecksumLine, line))
	}

	entry := checksumEntry{
		hash:     fields[0],
		filename: "-",
	}
	if len(fields) == maxChecksumFields {
		entry.filename = fields[1]
	}
	{
		algo, err := checksumAlgorithmForHash(entry.hash)
		if err != nil {
			return checksumEntry{}, err
		}
		entry.checksumAlgorithm = algo
	}
	return entry, nil
}

func (p *checksumParser) parseBSDStyleChecksum(_ context.Context, line string) checksumEntry {
	match := bsdStyleChecksums.FindStringSubmatch(line)
	return checksumEntry{
		hash:              match[3],
		filename:          match[2],
		checksumAlgorithm: checksumAlgorithm(match[1]),
	}
}

type checksumAlgorithm string

const (
	checksumAlgorithmSHA1   checksumAlgorithm = "SHA1"
	checksumAlgorithmSHA224 checksumAlgorithm = "SHA224"
	checksumAlgorithmSHA256 checksumAlgorithm = "SHA256"
	checksumAlgorithmSHA384 checksumAlgorithm = "SHA384"
	checksumAlgorithmSHA512 checksumAlgorithm = "SHA512"
)

const (
	bitsPerByte  = 8
	sha1LenBytes = 160 / bitsPerByte
)

func (a checksumAlgorithm) bytesLen() int {
	if a == checksumAlgorithmSHA1 {
		return sha1LenBytes
	}

	i, err := strconv.Atoi(strings.TrimPrefix(string(a), "SHA"))
	if err != nil {
		panic(err)
	}
	return i / bitsPerByte
}

func (a checksumAlgorithm) newDigest() hash.Hash {
	switch a {
	case checksumAlgorithmSHA1:
		return sha1.New() //nolint:gosec
	case checksumAlgorithmSHA224:
		return sha256.New224()
	case checksumAlgorithmSHA256:
		return sha256.New()
	case checksumAlgorithmSHA384:
		return sha512.New384()
	case checksumAlgorithmSHA512:
		return sha512.New()
	}
	panic("unexpected checksum algorithm: " + a)
}

func checksumAlgorithmForHash(hash string) (checksumAlgorithm, error) {
	algs := []checksumAlgorithm{
		checksumAlgorithmSHA1, checksumAlgorithmSHA224, checksumAlgorithmSHA256,
		checksumAlgorithmSHA384, checksumAlgorithmSHA512,
	}
	for _, alg := range algs {
		if alg.bytesLen()*2 == len(hash) {
			return alg, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrUnknownChecksumAlgorithm, hash)
}

type checksumEntry struct {
	checksumAlgorithm
	hash     string
	filename string
}

func (e checksumEntry) verify(asset githubapi.Asset, dest string) error {
	dig := e.newDigest()
	fp := path.Join(dest, asset.Name)
	var reader io.Reader
	f, err := os.Open(fp)
	if err != nil {
		return unexpected(err)
	}
	defer f.Close()
	reader = bufio.NewReader(f)
	if _, err = io.Copy(dig, reader); err != nil {
		return unexpected(err)
	}
	actual := hex.EncodeToString(dig.Sum(nil))
	if actual != e.hash {
		return fmt.Errorf("%w: %s, %s != %s",
			ErrChecksumMismatch, asset.Name, actual, e.hash)
	}
	return nil
}

func (e checksumEntry) Matches(name string) bool {
	return e.filename == "-" || e.filename == name
}

type checksumVerifier struct {
	entries []checksumEntry
}

func (c checksumVerifier) verify(
	ctx context.Context, assets []githubapi.Asset,
	dirFn func(curr githubapi.Asset) string,
) error {
	widgets := tui.WidgetsFrom(ctx)
	for _, entry := range c.entries {
		for i, curr := range assets {
			if entry.Matches(curr.Name) {
				spin := widgets.NewSpinner(ctx, fmt.Sprintf("🔍 Verifying checksum for %s",
					color.Cyan.Sprintf(curr.Name)))
				if err := spin.With(func(_ tui.Spinner) error {
					dest := dirFn(curr)
					return entry.verify(curr, dest)
				}); err != nil {
					return err
				}
				assets = append(assets[:i], assets[i+1:]...)
				break
			}
		}
	}

	if len(assets) > 0 {
		return errors.WithStack(fmt.Errorf("%w: %q", ErrNotVerifiedAssets, assets))
	}

	return nil
}
