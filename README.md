# Chaos Downloader

A simple tool to download, organize, and consolidate data from Project Discovery's Chaos dataset.
- - -
## Getting Started

Ensure Go is installed on your system to run this script.
- - -
### Steps

1. Clone this repository:

```
git clone https://github.com/tonytsep/chaos_downloader.git
```

2. Navigate to the repository folder:

```
cd chaos_downloader
```

3. Execute the script:

```
go run chaos_downloader.go
```

The script downloads ZIP files listed in https://chaos-data.projectdiscovery.io/index.json, extracts them into named directories, and compiles text from those directories into a single file named `everything.txt`.
- - -

## Usage

Running the script processes the data and generates `everything.txt` in the script's execution directory, alongside a folder `AllChaosData` containing the organized data.
- - -

## Contributing

Feel free to fork, modify, or suggest improvements to this script. Any contribution is welcome.
- - -

## License

This project is open-source, licensed under the MIT License.
