# Bulk File Renamer

Takes a CSV and renames files based on the source, column 0, to the destination, column 1. An optional prefix can be set for both the source and destination.

## Usage

```bash
renamefiles --csv <file>.csv --src-base /test1 -dst-base /test2

The input CSV file can also be supplied as a wildcard, e.g. *.csv

Options:
  --csv string
        The CSV file containing the source and destination file names. You can also use a wildcard, e.g. *.csv
  --dst-base string
        The base path for the destination files.
  --src-base string
        The base path for the source files.
  --copy-only bool
        Copy the files instead of moving them. (default false = files are moved)
```
