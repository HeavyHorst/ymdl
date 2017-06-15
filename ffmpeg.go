package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/cheggaaa/pb"
	"github.com/pkg/errors"
)

var errInvalidInput = errors.New("invalid duration")

func getLength(inputFile string) (int, error) {
	re := regexp.MustCompile("Duration: [0-9]+:[0-9]+:[0-9]+")

	cmd := exec.Command("ffmpeg", "-i", inputFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil && err.Error() != "exit status 1" {
		return 0, err
	}

	return durToSec(strings.TrimPrefix(re.FindString(string(stdoutStderr)), "Duration: "))
}

func transcode(inputFile, outputFile string, start, length float64) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile, "-ss", fmt.Sprintf("%.0f", start),
		"-t", fmt.Sprintf("%.0f", length),
		"-codec:a", "libmp3lame", "-qscale:a", "3",
		outputFile)
	return cmd.Run()
}

func convertTrack(inputFile string, mbr musicBrainzRecording, dlFolder string) error {
	l, err := getLength(inputFile)
	if err != nil {
		return errors.Wrap(err, "getLength failed")
	}
	artists := strings.Join(mbr.trackArtists, ",")
	trackName := fmt.Sprintf("%.2d %s - %s", mbr.trackNum, artists, mbr.trackTitle)
	trackFullPath := filepath.Join(dlFolder, trackName) + "." + "mp3"

	if err := transcode(inputFile, trackFullPath, 0.0, float64(l)); err != nil {
		return errors.Wrap(err, "transcode failed")
	}

	err = tagFile(trackFullPath, mbr.trackArtists, mbr.trackTitle, mbr.year, mbr.albumTitle, int(mbr.trackNum), int(mbr.trackCount), mbr.cdNum)
	if err != nil {
		return errors.Wrap(err, "tagFile failed")
	}
	return nil
}

func extractTracks(inputFile string, tracks []float64, mbr musicBrainzRelease, dlFolder string) error {
	l, err := getLength(inputFile)
	if err != nil {
		return errors.Wrap(err, "getLength failed")
	}

	tracks = append(tracks, float64(l))

	bar := pb.StartNew(len(mbr.tracks))
	defer bar.Finish()

	for i := 0; i < len(tracks)-1; i++ {
		if len(mbr.tracks) > i {

			artist := getArtists(mbr.tracks[i].Recording.ArtistCredit.NameCredits)
			title := mbr.tracks[i].Recording.Title

			trackName := fmt.Sprintf("%.2d %s - %s", i+1, strings.Join(artist, ","), title)
			trackFullPath := filepath.Join(dlFolder, trackName) + "." + "mp3"
			start := tracks[i]
			end := tracks[i+1]
			length := end - start

			if err := transcode(inputFile, trackFullPath, start, length); err != nil {
				return errors.Wrap(err, "transcode failed")
			}

			err = tagFile(trackFullPath, artist, title, mbr.year, mbr.title, i+1, len(mbr.tracks), -1)
			if err != nil {
				return errors.Wrap(err, "tagFile failed")
			}

			bar.Increment()
		}
	}
	return nil
}

func tagFile(path string, artists []string, title, year, album string, trackNum, tracksTotal int, cdNum int64) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return errors.Wrap(err, "id3v2 open failed")
	}
	defer tag.Close()

	tag.SetArtist(strings.Join(artists, "/"))
	tag.SetTitle(title)
	tag.SetYear(year)
	tag.SetAlbum(album)

	trckFrame := id3v2.TextFrame{
		Encoding: id3v2.ENUTF8,
		Text:     fmt.Sprintf("%d/%d", trackNum, tracksTotal),
	}
	tag.AddFrame(tag.CommonID("TRCK"), trckFrame)

	if cdNum >= 0 {
		tposFrame := id3v2.TextFrame{
			Encoding: id3v2.ENUTF8,
			Text:     fmt.Sprintf("%d", cdNum),
		}
		tag.AddFrame(tag.CommonID("TPOS"), tposFrame)
	}

	// Write it to file.
	if err = tag.Save(); err != nil {
		return errors.Wrap(err, "tag save failed")
	}

	return nil
}
