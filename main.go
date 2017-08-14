package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"os"

	"github.com/michiwend/gomusicbrainz"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/otium/ytdl"
	"github.com/pkg/errors"
	"github.com/subosito/norma"
)

const appName = "ymdl"
const version = "0.0.1-beta"
const contactURL = "github.com/HeavyHorst/ymdl"

func handleError(err error) {
	var e error
	if err != nil {
		if err != errNoRelease {
			e = errors.WithStack(err)
			fmt.Printf("%+v", e)
		} else {
			e = err
			fmt.Printf("%s\n", e.Error())
		}
		os.Exit(1)
	}
}

func dlRelease(library, url string, client *gomusicbrainz.WS2Client, vid *ytdl.VideoInfo) error {
	mbr, err := getAlbumInfo(client, vid.Title)
	handleError(err)

	dlFolder := filepath.Join(library, norma.Sanitize(mbr.artist), norma.Sanitize(mbr.title))
	dlFile := filepath.Join(dlFolder, norma.Sanitize(mbr.title))
	defer os.Remove(dlFile)

	tracks, err := getTracks(url)
	handleError(err)

	fmt.Println("\nDownloading Video:")
	err = download(vid, dlFile)
	handleError(err)

	fmt.Println("\nExtracting tracks:")
	return extractTracks(dlFile, tracks, mbr, dlFolder)
}

func dlRecord(library string, client *gomusicbrainz.WS2Client, vid *ytdl.VideoInfo) error {
	mbr, err := getTrackInfo(client, vid.Title)
	handleError(err)

	dlFolder := filepath.Join(library, norma.Sanitize(mbr.albumArtist), norma.Sanitize(mbr.albumTitle))
	dlFile := filepath.Join(dlFolder, norma.Sanitize(mbr.trackTitle))
	defer os.Remove(dlFile)

	fmt.Println("\nDownloading Video:")
	err = download(vid, dlFile)
	handleError(err)

	return convertTrack(dlFile, mbr, dlFolder)
}

func main() {
	var defaultLibraryPath string
	homeDir, err := homedir.Dir()
	if err != nil {
		//couldn't find the home directory
		defaultLibraryPath = "Music"
	} else {
		defaultLibraryPath = filepath.Join(homeDir, "Music")
	}

	dlTrack := flag.Bool("track", false, "download a single track from youtube")
	dlAlbum := flag.Bool("album", true, "download a complete album from youtube")
	printVersion := flag.Bool("version", false, "print the version and quit")
	libraryPath := flag.String("lib", defaultLibraryPath, "the path to your music library")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [album1 album2 ... albumN]\n\nParameters:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	client, err := gomusicbrainz.NewWS2Client("https://musicbrainz.org/ws/2", appName, version, contactURL)
	handleError(err)

	for _, url := range flag.Args() {
		vid, err := ytdl.GetVideoInfo(url)
		handleError(err)

		if *dlTrack {
			handleError(dlRecord(*libraryPath, client, vid))
		} else if *dlAlbum {
			handleError(dlRelease(*libraryPath, url, client, vid))
		}
	}
}
