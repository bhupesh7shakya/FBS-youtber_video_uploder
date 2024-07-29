package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func loadClientSecret(fileName string) ([]byte, error) {
	cs, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error reading client secret file: %w", err)
	}
	return cs, nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code:\n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	err = os.MkdirAll(tokenCacheDir, 0700)
	if err != nil {
		return "", err
	}
	return filepath.Join(tokenCacheDir, url.QueryEscape("youtube-go-quickstart.json")), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func handleError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func uploadVideo(youtuber *youtube.Service, filePath string, title string, description string, tags []string) {
	file, err := os.Open(filePath)
	if err != nil {
		handleError(err, "Error opening video file")
	}
	defer file.Close()

	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			// Tags:        tags,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "unlisted", // Options: public, unlisted, private
			MadeForKids:   false,
		},
	}

	call := youtuber.Videos.Insert([]string{"snippet", "status", "contentDetails"}, video)

	res, err := call.Media(file).Do()
	if err != nil {
		handleError(err, "Error uploading video")
	}
	fmt.Printf("Video Privacy Status: %s\n", res.Status.PrivacyStatus)
	fmt.Printf("Video uploaded successfully! Video ID: %s\n", res.Id)
}

func main() {
	ctx := context.Background()

	clientSecret, err := loadClientSecret("client_secret.json")
	if err != nil {
		handleError(err, "Error loading client secret")
	}

	// Update the scope to include upload permissions
	config, err := google.ConfigFromJSON(clientSecret, youtube.YoutubeScope, youtube.YoutubeUploadScope)
	if err != nil {
		handleError(err, "Error creating OAuth2 config")
	}

	client := getClient(ctx, config)

	youtuber, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		handleError(err, "Error creating YouTube service")
	}

	// Upload a video
	// videoPath := "C:\\Users\\bhupe\\Videos\\2024-07-20 11-21-12.mkv"
	// title := "Your Video Title"
	// description := "Your video description"
	// tags := []string{"tag1", "tag2"}

	// youtuber.Playlists.List()
	// uploadVideo(youtuber, videoPath, title, description, tags)
}
