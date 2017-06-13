/*
 * Copyright (c) 2014 Michael Wendland
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 *
 * 	Authors:
 * 		Michael Wendland <michael@michiwend.com>
 */

package gomusicbrainz

// CDStub represents an anonymously submitted track list.
type CDStub struct {
	ID        string `xml:"id,attr"` // seems not to be a valid MBID (UUID)
	Title     string `xml:"title"`
	Artist    string `xml:"artist"`
	Barcode   string `xml:"barcode"`
	Comment   string `xml:"comment"`
	TrackList struct {
		Count int `xml:"count,attr"`
	} `xml:"track-list"`
}

// SearchCDStub queries MusicBrainz´ Search Server for CDStubs.
//
// Possible search fields to provide in searchTerm are:
//
//	artist   artist name
//	title    release name
//	barcode  release barcode
//	comment  general comments about the release
//	tracks   number of tracks on the CD stub
//	discid   disc ID of the CD
//
// With no fields specified searchTerm searches only the artist Field. For more
// information visit
// https://musicbrainz.org/doc/Development/XML_Web_Service/Version_2/Search#CDStubs
func (c *WS2Client) SearchCDStub(searchTerm string, limit, offset int) (*CDStubSearchResponse, error) {

	result := cdStubListResult{}
	err := c.searchRequest("/cdstub", &result, searchTerm, limit, offset)

	rsp := CDStubSearchResponse{}
	rsp.WS2ListResponse = result.CDStubList.WS2ListResponse
	rsp.Scores = make(ScoreMap)

	for i, v := range result.CDStubList.CDStubs {
		rsp.CDStubs = append(rsp.CDStubs, v.CDStub)
		rsp.Scores[rsp.CDStubs[i]] = v.Score
	}

	return &rsp, err
}

// CDStubSearchResponse is the response type returned by the SearchCDStub method.
type CDStubSearchResponse struct {
	WS2ListResponse
	CDStubs []*CDStub
	Scores  ScoreMap
}

// ResultsWithScore returns a slice of CDStubs with a min score.
func (r *CDStubSearchResponse) ResultsWithScore(score int) []*CDStub {
	var res []*CDStub
	for _, v := range r.CDStubs {
		if r.Scores[v] >= score {
			res = append(res, v)
		}
	}
	return res
}

type cdStubListResult struct {
	CDStubList struct {
		WS2ListResponse
		CDStubs []struct {
			*CDStub
			Score int `xml:"http://musicbrainz.org/ns/ext#-2.0 score,attr"`
		} `xml:"cdstub"`
	} `xml:"cdstub-list"`
}
