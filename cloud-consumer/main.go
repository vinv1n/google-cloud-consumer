package main

import (
	"context"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"

	websocket "github.com/gorilla/websocket"
	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func main() {
	http.HandleFunc("/api", handleResponse)
	http.HandleFunc("/", landingPage)

	fmt.Println(http.ListenAndServe(":8081", nil))
}

func landingPage(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Server running :)")
}

func createGoogleCloudRequest(audio []byte, channel chan []byte) {
	response, err := makeRequest(audio)
	if err != nil {
		log.Panic("Error during request to google api", err)

	}
	channel <- response
	close(channel)
}

func createWebsocket(res []byte) {
	// creates new websocket
	conn, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:8888", nil)
	if err != nil {
		log.Panic("Error during connecting websocket", err)
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, res)
	if err != nil {
		log.Panic("Failed send data to the socket", err)
	}
	// close the socket
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func handlePost(request *http.Request) {
	audio, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Panic("Error during reading request content. Error ", err)
	}

	channel := make(chan []byte, 100)

	// makes request to the google cloud api
	go createGoogleCloudRequest(audio, channel)
	defer createWebsocket(<-channel)
}

func handleResponse(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		handlePost(request)
	default:
		http.Error(writer, "Method not supported", 403)
	}
}

func makeRequest(audio []byte) ([]byte, error) {
	// Based on example provided on official google cloud documentation
	// TODO investigate if we could use websockets from mic thingy to stream
	// audio to cloud

	// create client for google cloud
	context := context.Background()
	client, err := speech.NewClient(context)
	if err != nil {
		log.Fatalf("Could not create client for google cloud, error %v", err)
		return nil, err
	}

	response, err := client.Recognize(context, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: 16000,
			LanguageCode:    "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: audio},
		},
	})
	if err != nil {
		log.Fatalf("Failed to recognized audio %s", err)
		return nil, err
	}
	results := parseResults(response)
	return results, nil
}

func parseResults(resp *speechpb.RecognizeResponse) []byte {
	// get the alternative that have highest confidence

	var confidence float32;
	var transcript string;
	for _, res := range resp.Results {
		for _, alt := range res.Alternatives {
			if confidence == 0.0 || alt.Confidence >= confidence {
				transcript = alt.Transcript
				confidence = alt.Confidence
			}
		} 
	}
	return []byte(transcript)
}