# simple-to-boost

Go CLI tool to import a Simplenote bulk export into Boostnote.

## Before you start

Make sure to back up your existing Boost files before running this. I whipped up this quick project for my own use. If you want to help contribute to it, that'd be great 🙌

## Install

Use `dep`, or `go get github.com/scottrigby/simple-to-boost`.

## Usage

1. Export your Simplenote data by following [these instructions](https://simplenote.com/help/#export). Unzip the file and copy the backup directory location.
1. Note your Boostnote storage directory (the default is `~/Boostnote`, but users may create other storage locations). See the [data format](https://github.com/BoostIO/Boostnote/wiki/Data-format) Wiki page for more info.
1. Run `simple-to-boost`, and you will be prompted for:
    - Simplenote export directory (paste the directory location from step 1 above. You may use `~` for home directory expansion)
    - Boost storage directory (defaults to `~/Boostnote`, which will work if that directory exists. If not, locate your preferred Boost storage directory)
    - Select folder (you will be given the option to select existing folders for your desired storage directory, or to automatically create a new one)
1. You should see the message:
    > Imported! Quit and reopen Boost to see your files.

    Your newly imported Boost files will retain the Simplenote updated date metadata, so they should still be in the correct order!
