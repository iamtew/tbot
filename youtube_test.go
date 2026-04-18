package main

import (
	"testing"
)

func TestFetchYouTubeMetadata(t *testing.T) {
	barrel := NewYouTubeBarrel()
	testLinks := []string{
		"https://youtu.be/BbNl62r1c18",
		"https://youtu.be/DtYTxGK9Pds",
		"https://www.youtube.com/watch?v=7D-aJhNqegY",
		// Add more test links here as needed
	}

	for _, link := range testLinks {
		title, likes, dislikes, uploadDate, channelTitle, err := barrel.fetchYouTubeMetadata(t.Logf, link)
		if err != nil {
			t.Errorf("Error fetching metadata for %s: %v", link, err)
		} else {
			t.Logf("Link: %s\nTitle: %s\nLikes: %s\nDislikes: %s\nUpload: %s\nChannel: %s\n", link, title, likes, dislikes, uploadDate, channelTitle)
		}
	}
}
