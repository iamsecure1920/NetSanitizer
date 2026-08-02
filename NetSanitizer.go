// NetSanitizer — URL deduplication for reconnaissance output.
//
// Archive sources (gau, waybackurls, waymore) return the same endpoint hundreds
// of times with different values in the same parameters. What a tester needs is
// the set of *distinct injection points*, not the set of distinct URLs. On a
// real engagement 5,065 archived URLs reduced to 61 distinct injection points;
// everything between those numbers is noise that costs scan budget.
//
//	cat urls.txt | netsanitizer
//	netsanitizer urls.txt
//	gau example.com | netsanitizer -keep-assets
//
// Deduplication key: scheme + host + path + the *set of parameter names*.
// Values are ignored — /item?id=1 and /item?id=999 are one injection point —
// but /item?id=1 and /item?ref=x are two, because they take different input.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
)

// Static assets: no server-side input handling, so nothing to inject into.
// Note that .json and .xml are deliberately absent — an API returning JSON is
// among the most interesting things recon finds, and dropping
// /api/v1/users.json?id=1 discards a prime IDOR candidate. The original tool
// listed .xml in both the ignore and web lists, where the ignore check ran
// first and silently won.
var assetSuffixes = []string{
	"css", "gif", "jpg", "png", "jpeg", "svg", "ico", "webp",
	"otf", "ttf", "woff", "woff2", "eot", "swf",
	"zip", "gz", "tar", "pdf", "doc", "docx", "ppt", "pptx", "xls", "xlsx",
	"ogg", "mp4", "mp3", "mov", "avi", "webm",
}

// JavaScript is an asset for scanning purposes but a target for analysis: it
// carries endpoints, API keys and internal hostnames. Kept behind a flag so a
// JS-analysis pipeline can ask for it explicitly.
var scriptSuffixes = []string{"js", "mjs", "jsx", "map"}

type entry struct {
	raw    string
	params int
}

func normalize(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	u.Fragment = "" // #frag never reaches the server
	q := u.Query()
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := url.Values{}
	for _, k := range keys {
		sorted[k] = q[k]
	}
	u.RawQuery = sorted.Encode()
	return u.String(), nil
}

func hasSuffix(p string, suffixes []string) bool {
	lower := strings.ToLower(p)
	for _, s := range suffixes {
		if strings.HasSuffix(lower, "."+s) {
			return true
		}
	}
	return false
}

// key identifies an injection point: where the request goes, and which inputs
// it accepts. Parameter *values* are excluded on purpose.
func key(u *url.URL) string {
	names := make([]string, 0, len(u.Query()))
	for k := range u.Query() {
		names = append(names, k)
	}
	sort.Strings(names)
	path := u.Path
	if path == "" {
		path = "/"
	}
	return u.Scheme + "://" + u.Host + path + "?" + strings.Join(names, "&")
}

func run(in io.Reader, out io.Writer, keepAssets, keepScripts bool) (kept, seen int) {
	best := map[string]entry{}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024) // archive lines get long

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		seen++

		normalized, err := normalize(line)
		if err != nil {
			continue
		}
		u, err := url.Parse(normalized)
		if err != nil || u.Host == "" {
			continue
		}

		if !keepAssets && hasSuffix(u.Path, assetSuffixes) {
			continue
		}
		if !keepScripts && hasSuffix(u.Path, scriptSuffixes) {
			continue
		}

		k := key(u)
		// Both sides compared on the normalized form. The original compared a
		// raw stored URL against a normalized candidate, so the "more
		// parameters wins" rule was measuring two different things.
		candidate := entry{raw: normalized, params: len(u.Query())}
		if existing, ok := best[k]; !ok || candidate.params > existing.params {
			best[k] = candidate
		}
	}

	// Sorted output. Go randomises map iteration, so the original printed a
	// different order every run — which makes diffing two recon passes
	// impossible, and diffing is the point of running recon twice.
	keys := make([]string, 0, len(best))
	for k := range best {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintln(out, best[k].raw)
	}
	return len(best), seen
}

func main() {
	keepAssets := flag.Bool("keep-assets", false, "keep static assets (images, fonts, archives)")
	keepScripts := flag.Bool("keep-scripts", false, "keep .js and source maps — useful before JS analysis")
	quiet := flag.Bool("q", false, "suppress the summary on stderr")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "netsanitizer — collapse recon URLs to distinct injection points\n\n")
		fmt.Fprintf(os.Stderr, "  netsanitizer [flags] [file]\n")
		fmt.Fprintf(os.Stderr, "  cat urls.txt | netsanitizer\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	in := io.Reader(os.Stdin)
	if flag.NArg() == 1 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %v\n", flag.Arg(0), err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	} else if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}

	kept, seen := run(in, os.Stdout, *keepAssets, *keepScripts)
	if !*quiet {
		fmt.Fprintf(os.Stderr, "netsanitizer: %d urls -> %d distinct injection points\n", seen, kept)
	}
}
