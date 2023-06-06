package download

import (
	"bufio"
	"context"
	"crypto/sha1"
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

	githubapi "github.com/cardil/ghet/pkg/github/api"
	"github.com/cardil/ghet/pkg/output"
	"github.com/cardil/ghet/pkg/output/tui"
	slog "github.com/go-eden/slf4go"
	"github.com/gookit/color"
	"github.com/pkg/errors"
)

// ErrTooManyChecksums is returned when there are more than one checksums.
var ErrTooManyChecksums = errors.New("too many checksums")

// ErrUnknownChecksumAlgorithm is returned when the checksum algorithm is unknown.
var ErrUnknownChecksumAlgorithm = errors.New("unknown checksum algorithm")

// ErrChecksumMismatch is returned when the checksum does not match.
var ErrChecksumMismatch = errors.New("checksum mismatch")

// ErrNotVerifiedAssets is returned when there are no verified assets.
var ErrNotVerifiedAssets = errors.New("not verified assets")

var bsdStyleChecksums = regexp.MustCompile(`^(SHA[0-9]{1,3})\s+\([^)]+\)\s+=\s+([a-fA-F0-9]{32,128})$`)

func verifyChecksums(ctx context.Context, assets []githubapi.Asset, args Args) error {
	l := output.LoggerFrom(ctx)
	widgets := tui.WidgetsFrom(ctx)
	index := githubapi.CreateIndex(assets)
	if len(index.Checksums) == 0 {
		l.Debug("No checksums to verify")
		widgets.Printf(ctx, "🕵 No checksums to verify")
		return nil
	}

	ca := index.Checksums[0]

	if len(index.Checksums) > 1 {
		options := make([]string, len(index.Checksums))
		for i, c := range index.Checksums {
			options[i] = c.Name
		}
		intwidgets, err := widgets.Interactive(ctx)
		if err != nil {
			if errors.Is(err, tui.ErrNotInteractive) {
				l.Errorf("Number of checksums is %d. Expected just one.", len(index.Checksums))
				return fmt.Errorf("%w: %d", ErrTooManyChecksums, len(index.Checksums))
			}
			return fmt.Errorf("%w: %v", ErrUnexpected, err)
		}
		selected := intwidgets.Ask(ctx, options,
			"⚠️ More than one checksum file found. Choose proper one")
		for _, c := range index.Checksums {
			if c.Name == selected {
				ca = c
				break
			}
		}
	}

	l = l.WithFields(slog.Fields{"checksum": ca.Name})
	l.Debug("Verifying checksum")

	csp := checksumParser{Asset: ca, Args: args}
	if cs, err := csp.parse(ctx); err != nil {
		return err
	} else {
		err = cs.verify(ctx, append(index.Archives, index.Other...), args.Destination)
		if err != nil {
			return err
		}
	}

	widgets.Printf(ctx, "✅ All checksums match the downloaded assets")

	l.Debugf("Deleting the checksums file(s): %q", index.Checksums)
	for _, c := range index.Checksums {
		if err := os.Remove(path.Join(args.Destination, c.Name)); err != nil {
			return unexpected(err)
		}
	}

	return nil
}

type checksumParser struct {
	githubapi.Asset
	Args
	*checksums
}

func (p *checksumParser) parse(ctx context.Context) (*checksums, error) {
	l := output.LoggerFrom(ctx)
	fp := path.Join(p.Destination, p.Name)
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
	p.checksums = &checksums{
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

	return p.checksums, nil
}

func (p *checksumParser) parseLine(ctx context.Context, line string) error {
	var entry checksumEntry
	if bsdStyleChecksums.MatchString(line) {
		entry = p.parseBSDStyleChecksum(ctx, line)
	} else {
		if e, err := p.parseRegularChecksum(line); err != nil {
			return err
		} else {
			entry = e
		}
	}
	p.checksums.entries = append(p.checksums.entries, entry)
	return nil
}

func (p *checksumParser) parseRegularChecksum(line string) (checksumEntry, error) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return checksumEntry{}, unexpected(fmt.Errorf("invalid checksum line: %s", line))
	}

	entry := checksumEntry{
		hash:     fields[0],
		filename: fields[1],
	}
	if algo, err := checksumAlgorithmForHash(entry.hash); err != nil {
		return checksumEntry{}, err
	} else {
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

func (a checksumAlgorithm) bytesLen() int {
	if a == checksumAlgorithmSHA1 {
		return 20
	}

	i, err := strconv.Atoi(strings.TrimPrefix(string(a), "SHA"))
	if err != nil {
		panic(err)
	}
	return i / 8
}

func (a checksumAlgorithm) newDigest() hash.Hash {
	switch a {
	case checksumAlgorithmSHA1:
		return sha1.New()
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

func (e checksumEntry) verify(_ context.Context, asset githubapi.Asset, dest string) error {
	dig := e.newDigest()
	fp := path.Join(dest, asset.Name)
	var reader io.Reader
	if f, err := os.Open(fp); err != nil {
		return unexpected(err)
	} else {
		defer f.Close()
		reader = bufio.NewReader(f)
	}
	if _, err := io.Copy(dig, reader); err != nil {
		return unexpected(err)
	}
	actual := hex.EncodeToString(dig.Sum(nil))
	if actual != e.hash {
		return fmt.Errorf("%w: %s, %s != %s",
			ErrChecksumMismatch, asset.Name, actual, e.hash)
	}
	return nil
}

type checksums struct {
	entries []checksumEntry
}

func (c checksums) verify(ctx context.Context, assets []githubapi.Asset, dest string) error {
	widgets := tui.WidgetsFrom(ctx)
	for _, entry := range c.entries {
		for i, curr := range assets {
			if entry.filename == curr.Name {
				spin := widgets.NewSpinner(ctx, fmt.Sprintf("🔍 Verifying checksum for %s",
					color.Cyan.Sprintf(curr.Name)))
				if err := spin.With(func(_ tui.Spinner) error {
					if err := entry.verify(ctx, curr, dest); err != nil {
						return err
					}
					return nil
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
