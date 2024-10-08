# NetSanitizer

NetSanitizer is a URL deduplication tool designed to filter and clean a list of URLs by normalizing them, removing duplicates, and ignoring certain file types. It ensures a clean set of URLs, prioritizing those with more query parameters for web-related paths. This tool is handy for bug bounty hunters and penetration testers during reconnaissance.

## Features

- **URL Normalization**: Removes fragments and sorts query parameters for consistency.
- **File Type Filtering**: Ignores URLs with specified file extensions (e.g., images, scripts, documents).
- **Path Deduplication**: Deduplicates URLs based on their paths, prioritizing more informative URLs.
- **Web Suffix Handling**: Recognizes and processes common web-related file extensions.


## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/NetSanitizer.git
    cd NetSanitizer
    ```

2. Build the executable:
    ```sh
    go build  NetSanitizer
    ```

## Usage

Provide a file containing URLs, and NetSanitizer will output a deduplicated list to the console.

```sh
./NetSanitizer <input_file>
 ```

Example
Consider you have a file urls.txt with the following content:

bash
```sh
http://example.com/path?b=2&a=1
http://example.com/path?b=2&a=1#fragment
http://example.com/path2
http://example.com/image.png
```
Running ./NetSanitizer urls.txt will produce:

bash
```sh
http://example.com/path?b=2&a=1
http://example.com/path2
```

https://github.com/user-attachments/assets/a1e93aaf-c669-4288-8d81-c149cb8f87b7
