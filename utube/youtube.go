package utube

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

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
	openInBrowser(authURL)
	var code string
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code = r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Authorization code received. You can close this window.")
		// Stop the server once the code is received
		// Use a global variable to signal when to stop the server
		stopServer()
	})

	// Start the HTTP server
	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait until the code is received
	waitForCode()

	// Exchange the authorization code for a token
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Global variable to stop the server
var stop = make(chan struct{})

func waitForCode() {
	<-stop
}

// Function to stop the HTTP server
func stopServer() {
	close(stop)
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
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentialssss")
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

func UploadVideo(youtuber *youtube.Service, filePath string, title string, description string, tags []string) string {
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
	return res.Id
}

// GetPlaylistIDByName returns the ID of the playlist with the specified name.
func GetPlaylistIDByName(youtuber *youtube.Service, playlistName string) (string, error) {
	call := youtuber.Playlists.List([]string{"snippet"}).
		Mine(true)

	playlistListResponse, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("Error retrieving playlists: %v", err)
	}

	for _, playlist := range playlistListResponse.Items {
		if playlist.Snippet.Title == playlistName {
			return playlist.Id, nil
		}
	}

	return "", fmt.Errorf("Playlist with name '%s' not found", playlistName)
}

func CreatePlaylist(youtuber *youtube.Service, playlistTitle string, description string) (string, error) {
	existingPlaylistID, err := GetPlaylistIDByName(youtuber, playlistTitle)
	if err == nil {
		fmt.Printf("Playlist already exists! Playlist ID: %s\n", existingPlaylistID)
		return existingPlaylistID, nil
	}
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlistTitle,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public", // Options: public, unlisted, private
		},
	}

	call := youtuber.Playlists.Insert([]string{"snippet", "status"}, playlist)

	createdPlaylist, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("Error creating playlist: %v", err)
	}

	fmt.Printf("Playlist created successfully! Playlist ID: %s\n", createdPlaylist.Id)
	return createdPlaylist.Id, nil
}

func openInBrowser(url string) error {
	url = strings.ReplaceAll(url, "&", "^&")

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
func AddVideoToPlaylist(youtuber *youtube.Service, playlistId string, videoId string) error {
	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistId,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoId,
			},
		},
	}

	call := youtuber.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err := call.Do()
	if err != nil {
		return fmt.Errorf("Error adding video to playlist: %v", err)
	}

	fmt.Printf("Video added to playlist successfully! Playlist ID: %s, Video ID: %s\n", playlistId, videoId)
	return nil
}

// func YoutubeSetup() *youtube.YoutubeService {
// 	ctx := context.Background()

// 	clientSecret, err := loadClientSecret("client_secret.json")
// 	if err != nil {
// 		handleError(err, "Error loading client secret")
// 	}

// 	// Update the scope to include upload permissions
// 	config, err := google.ConfigFromJSON(clientSecret, youtube.YoutubeScope, youtube.YoutubeUploadScope)
// 	if err != nil {
// 		handleError(err, "Error creating OAuth2 config")
// 	}

// 	client := getClient(ctx, config)

// 	youtuber, err := youtube.NewService(ctx, option.WithHTTPClient(client))
// 	if err != nil {
// 		handleError(err, "Error creating YouTube service")
// 	}
// 	return &youtuber

// 	// Upload a video
// 	// videoPath := "C:\\Users\\bhupe\\Videos\\2024-07-20 11-21-12.mkv"
// 	// title := "Your Video Title"
// 	// description := "Your video description"
// 	// tags := []string{"tag1", "tag2"}

// 	// uploadVideo(youtuber, videoPath, title, description, tags)
// }

func YouTubeSetup() (*youtube.Service, error) {
	ctx := context.Background()

	clientSecret, err := loadClientSecret("client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("Error loading client secret: %v", err)
	}

	// Update the scope to include upload permissions
	config, err := google.ConfigFromJSON(clientSecret, youtube.YoutubeScope, youtube.YoutubeUploadScope)
	if err != nil {
		return nil, fmt.Errorf("Error creating OAuth2 config: %v", err)
	}

	client := getClient(ctx, config)

	youtuber, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Error creating YouTube service: %v", err)
	}

	return youtuber, nil
}
