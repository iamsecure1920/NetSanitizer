# NetSanitizer

NetSanitizer collapses reconnaissance URL dumps into the set of **distinct
injection points** — what a tester actually needs — rather than the set of
distinct URLs.

Archive sources (gau, waybackurls, waymore) return the same endpoint hundreds of
times with different *values* in the same parameters. Measured on a real
engagement, 5,065 archived URLs contained 61 distinct injection points;
everything in between is scan budget spent re-testing one endpoint.

## Usage

```sh
netsanitizer urls.txt
cat urls.txt | netsanitizer
gau example.com | netsanitizer -keep-scripts
```

| flag | effect |
|---|---|
| `-keep-assets` | keep images, fonts, archives (dropped by default) |
| `-keep-scripts` | keep `.js` and source maps — useful before JS analysis |
| `-q` | suppress the summary on stderr |

## How it deduplicates

The key is `scheme + host + path + the set of parameter NAMES`. Values are
ignored, so `/item?id=1` and `/item?id=999` are one injection point — but
`/item?id=1` and `/item?ref=x` are two, because they accept different input.
Where several URLs share a key, the one with the most parameters wins.

Fragments are stripped (`#frag` never reaches the server) and query parameters
are sorted, so ordering differences do not create duplicates.

## What is kept

`.json` and `.xml` are **not** dropped. An API returning JSON is among the most
interesting things recon finds, and discarding `/api/v1/users.json?id=1` throws
away a prime IDOR candidate. Only genuinely static assets — images, fonts,
media, archives — are removed by default.

JavaScript is dropped by default but kept with `-keep-scripts`, because JS
bundles carry endpoints, API keys and internal hostnames that are worth
analysing separately.

## Output

Sorted, and therefore stable across runs. Two passes over the same input produce
byte-identical output, so recon results can be diffed to find what changed —
which is the point of running recon twice.

## Build

```sh
go build -o netsanitizer NetSanitizer.go
```

## Example

```
$ cat urls.txt
https://x.com/shop/item?id=1
https://x.com/shop/item?id=999
https://x.com/shop/cart?sess=9
https://x.com/api/v1/users.json?id=1
https://x.com/static/app.js
https://x.com/a/b?p=1#fragment

$ netsanitizer urls.txt
https://x.com/a/b?p=1
https://x.com/api/v1/users.json?id=1
https://x.com/shop/cart?sess=9
https://x.com/shop/item?id=1
netsanitizer: 6 urls -> 4 distinct injection points
```
