# simple-to-boost

Go CLI tool to import a Simplenote bulk export into Boostnote.

## Before you start

### This is for fun

Use this if you prefer CLI over UI.

Boostnote already allows you to [import .md/.txt files](https://github.com/BoostIO/Boostnote/wiki/md-txt-import) by drag and drop, and Simplenote exports notes as a zip of `.txt` files.

### Backup

Make sure to back up your existing Boost files before running this. I whipped up this quick project for my own use. If you want to help contribute to it for fun, I'm usually open to collaboration 🙌

## Installation

### Binaries (recommended)

Download your preferred asset from the [releases page](https://github.com/scottrigby/simple-to-boost/releases) and install manually.

### Homebrew (MacOS)

```console
$ brew install scottrigby/tap/simple-to-boost
```

### Go get (for contributing)

```console
$ go get -d github.com/scottrigby/simple-to-boost
$ cd $GOPATH/src/github.com/scottrigby/simple-to-boost
$ dep ensure -vendor-only
$ go install
```

## Usage

1. Export your Simplenote data by following [these instructions](https://simplenote.com/help/#export). Unzip the file and copy the backup directory location.
1. Note your Boostnote storage directory (the default is `~/Boostnote`, but users may create other storage locations). See the [data format](https://github.com/BoostIO/Boostnote/wiki/Data-format) Wiki page for more info.
1. Run `simple-to-boost`, and you will be prompted for:
    - Simplenote export directory (paste the directory location from step 1 above. You may use `~` for home directory expansion)
    - Boost storage directory (defaults to `~/Boostnote`, which will work if that directory exists. If not, locate your preferred Boost storage directory)
    - Select folder (you will be given the option to select existing folders for your desired storage directory, or to automatically create a new one)
    - You should see the message:
        > Imported! Quit and reopen Boost to see your files.

        Your newly imported Boost files will retain the Simplenote updated date metadata, so they should still be listed in the correct order in Boostnote.
