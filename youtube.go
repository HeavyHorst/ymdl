package main

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb"
	"github.com/marcmak/calc/calc"
	"github.com/otium/ytdl"
	"github.com/pkg/errors"
)

func getTracks(url string) ([]float64, error) {
	var tracks []float64

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return tracks, err
	}

	re := regexp.MustCompile("\\((.*?)\\)")
	// extract description and title
	desc := doc.Find("#eow-description > a")
	desc.Each(func(i int, s *goquery.Selection) {
		if attr, ok := s.Attr("onclick"); ok {
			tracks = append(tracks, calc.Solve(re.FindString(attr)))
		}
	})

	return tracks, nil
}

func getContentLength(url *url.URL) (int, error) {
	// try to get the length from the url query params
	if clString, ok := url.Query()["clen"]; ok {
		return strconv.Atoi(clString[0])
	}

	// try to get the length from the http header
	response, err := http.Head(url.String())
	if err != nil {
		return 0, errors.Wrap(err, "couldn't get content length from http header")
	}

	if response.StatusCode != http.StatusOK {
		return 0, errors.New("server returned non-200 status: " + response.Status)
	}

	return strconv.Atoi(response.Header.Get("Content-Length"))
}

func download(vid *ytdl.VideoInfo, outfile string) error {
	os.MkdirAll(filepath.Dir(outfile), 0777)
	// get best AudioBitrate
	var format ytdl.Format
	for _, v := range vid.Formats.Best(ytdl.FormatAudioBitrateKey) {
		if v.AudioBitrate > format.AudioBitrate {
			format = v
		}
	}

	// get downloadURL and ...
	dlURL, err := vid.GetDownloadURL(format)
	if err != nil {
		return errors.Wrap(err, "couldn't get download url")
	}
	//... get the content length
	clen, err := getContentLength(dlURL)
	if err != nil {
		return errors.Wrap(err, "getContentLength failed")
	}

	file, err := os.Create(outfile)
	if err != nil {
		return errors.Wrap(err, "couldn't create the video file")
	}
	defer file.Close()

	// create and start bar
	bar := pb.New(clen).SetUnits(pb.U_BYTES)
	bar.ShowTimeLeft = true
	bar.ShowSpeed = true
	bar.Start()
	defer bar.Finish()

	// create multi writer
	writer := io.MultiWriter(file, bar)

	err = vid.Download(format, writer)
	return errors.Wrap(err, "download failed")
}
