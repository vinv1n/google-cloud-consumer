package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

type File struct {
	audio   byte
	content string
	path    string
}

func main() {
	http.HandleFunc("/api", handleResponse)
	http.HandleFunc("/", landingPage)

	http.ListenAndServe("0.0.0.0:8081", nil)
}

func landingPage(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Running in ports 8081 :)")
}

func createGoogleCloudRequest(file File, channel chan bool) {
	/*
		A value of file should be used as long as possible as it is more efficient
		in golang that passing a pointer
	*/
	err, response := makeRequest(&file)
	if err != nil {
		log.Fatal("Error during request to google api", err)
		channel <- false

	} else {
		// TODO add stuff to File structure
		log.Fatal("success")

		// this could be done better and more elegant manner
		channel <- true
	}
	log.Fatal("Response", response)
	close(channel)
}

func parsePostRequest(resp *http.Request) File {
	// FIXME
	file := new(File)
	return *file
}

func handleResponse(writer http.ResponseWriter, request *http.Request) {

	switch request.Method {
	case "POST":
		file := parsePostRequest(request)
		channel := make(chan bool, 1)

		// creates a new go runtime routine
		go createGoogleCloudRequest(file, channel)

		// waits untill response is sent to the channel
		defer func() {
			// TODO fix the result codes and
			result := <-channel
			if result == true {
				fmt.Fprint(writer, "Success", 200)
			} else {
				fmt.Fprint(writer, "Failure", 400)
			}
		}()
	default:
		http.Error(writer, "Method not supported", 403) // FIXME the error code should be replaced
	}
}

func makeRequest(file *File) ([]string, error) {
	/*
		Based on example provided on official google cloud docmentation
		TODO add some better handling for errors
	*/

	// base context
	context := context.Background()

	// crate client for google cloud
	client, err := speech.NewClient(context)
	if err != nil {
		log.Fatalf("Could not create client for google cloud, error %v", err)
		return nil, err
	}

	audio, err := ioutil.ReadFile(file.path)
	if err != nil {
		log.Fatalf("Could not read file %v", err)
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

// TODO check type and content of response
// also rethink all of this
func parseResults(resp *speechpb.RecognizeResponse) []string {
	count := 0
	resmap := make([]string, len(resp.GetResults()))

	// HACK: google api returns float32, this is noway near clean and nice solution
	var confidence float32 = -1.0
	var word string = ""
	for _, res := range resp.Results {
		for _, alt := range res.Alternatives {
			if confidence == -1 || alt.Confidence > confidence {
				confidence = alt.Confidence
				word = alt.Transcript
			}
		}
		resmap = append(resmap, word)
		count += 1
	}
	return resmap

}
