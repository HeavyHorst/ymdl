package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"time"

	"io/ioutil"

	"github.com/michiwend/gomusicbrainz"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var errNoRelease = errors.New("couldn't find a release")

type musicBrainzRecording struct {
	year         string
	albumTitle   string
	cdNum        int64
	trackNum     int64
	trackCount   int64
	albumArtist  string
	trackArtists []string
	trackTitle   string
}

type musicBrainzRelease struct {
	artist string
	title  string
	year   string
	tracks []*gomusicbrainz.Track
}

func scanLines(scanTo map[string]*string) {
	for k, v := range scanTo {
		fmt.Printf("%s [%s]: ", k, *v)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		s := scanner.Text()
		if s != "" {
			*v = s
		}
	}
}

func getArtistAlbumOrTrack(query string) (artist string, release string) {
	info := strings.Split(query, "-")
	if len(info) > 1 {
		artist = strings.TrimSpace(info[0])
		release = strings.TrimSpace(info[1])
	}
	return
}

func getTrackInfo(client *gomusicbrainz.WS2Client, query string) (musicBrainzRecording, error) {
	var recording musicBrainzRecording
	artist, track := getArtistAlbumOrTrack(query)
	scanTo := map[string]*string{
		"Artist": &artist,
		"Track":  &track,
	}

	scanLines(scanTo)

	query = fmt.Sprintf(`artist:"%s" AND %s`, strings.TrimSpace(artist), strings.TrimSpace(track))

	req, err := http.NewRequest("GET", "https://musicbrainz.org/ws/2/recording/", nil)
	if err != nil {
		return recording, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("fmt", "json")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s ( %s )", appName, version, contactURL))

	fmt.Println("\nSearching track on musicbrainz: ")

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return recording, err
	}
	defer resp.Body.Close()

	json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return recording, errors.Wrap(err, "reading response body failed")
	}

	proceed := true
	value := gjson.GetBytes(json, "recordings")

	value.ForEach(func(key, value gjson.Result) bool {

		var artists []string
		value.Get("artist-credit.#.artist.name").ForEach(func(key, value gjson.Result) bool {
			artists = append(artists, value.String())
			return true
		})

		trackTitle := value.Get("title").String()
		value.Get("releases").ForEach(func(key, release gjson.Result) bool {
			release.Get("media").ForEach(func(key, media gjson.Result) bool {

				recording = musicBrainzRecording{
					albumArtist:  artist,
					trackArtists: artists,
					cdNum:        media.Get("position").Int(),
					trackCount:   media.Get("track-count").Int(),
					trackNum:     media.Get("track-offset").Int() + 1,
					albumTitle:   release.Get("title").String(),
					year:         release.Get("date").String(),
					trackTitle:   trackTitle,
				}

				if len(recording.year) >= 4 {
					recording.year = recording.year[0:4]
				}

				fmt.Printf("Release: %s (%s)\nFormat: %s\nTrack: %.2d/%.2d %s - %s\n",
					recording.albumTitle, recording.year, media.Get("format").String(), recording.trackNum, recording.trackCount, artists, recording.trackTitle)

				if askForConfirmation("Choose track?") {
					proceed = false
				}

				fmt.Println()
				return proceed
			})
			return proceed
		})
		return proceed
	})

	if proceed == false {
		return recording, nil
	}
	return recording, errNoRelease
}

func getAlbumInfo(client *gomusicbrainz.WS2Client, query string) (musicBrainzRelease, error) {
	var mbr musicBrainzRelease
	artist, release := getArtistAlbumOrTrack(query)
	scanTo := map[string]*string{
		"Artist":  &artist,
		"Release": &release,
	}

	scanLines(scanTo)

	query = fmt.Sprintf(`artist:"%s" AND %s`, strings.TrimSpace(artist), strings.TrimSpace(release))
	fmt.Println("\nSearching release on musicbrainz: ")
	resp, err := client.SearchRelease(query, 5, -1)
	if err != nil {
		return mbr, errors.Wrap(err, "SearchRelease failed")
	}

	for _, release := range resp.Releases {
		rec, _ := client.LookupRelease(release.Id(), "artist-credits", "labels", "discids", "recordings")
		var label string
		if len(rec.LabelInfos) > 0 {
			label = rec.LabelInfos[0].Label.Name
		}
		fmt.Printf("Label: %s \nRelease: %s (%d)\n", label, release.Title, release.Date.Year())
		for _, v := range rec.Mediums {
			fmt.Println("Format: " + v.Format)
			for _, t := range v.Tracks {
				fmt.Printf("\t%s: %s - %s\n", t.Number, getArtists(t.Recording.ArtistCredit.NameCredits), t.Recording.Title)
			}

			if askForConfirmation("Choose release?") {
				mbr.artist = artist
				mbr.title = release.Title
				mbr.year = strconv.Itoa(rec.Date.Year())
				mbr.tracks = v.Tracks
				return mbr, nil
			}
			fmt.Println()
		}
	}
	return mbr, errNoRelease
}

func getArtists(nc []gomusicbrainz.NameCredit) []string {
	artists := make([]string, 0, len(nc))
	for _, v := range nc {
		artists = append(artists, v.Artist.Name)
	}

	return artists
}
