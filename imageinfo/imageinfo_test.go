package imageinfo

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var testImage = &ImageInfo{
	Repo:      "library/golang",
	baseURL:   "https://registry.hub.docker.com",
	startTime: time.Now(),
}

func TestGetAuthToken(t *testing.T) {
	// expected length of token
	want := 2048

	err := testImage.getAuthToken()
	if err != nil {
		t.Errorf("getAuthToken(): %s", err.Error())
	}
	got := len(testImage.AuthToken)
	if got != want {
		t.Errorf("got: %d | want: %d", got, want)
	}
}
func TestGetTagsFail(t *testing.T) {
	err := testImage.getTags()
	if err != nil {
		errMsg := err.Error()
		if errMsg != "Could Not Find Image" {
			t.Errorf("got: %s | want: %s", errMsg, "Could Not Find Image")
		}
	}
}

func TestGetTagsSuccess(t *testing.T) {
	// TODO(Davy): This needs mocking better as is pretty shit
	want := []string{
		"1-alpine",
		"1-alpine3.5",
		"1-alpine3.6",
		"1-alpine3.7",
		"1-cross",
		"1-jessie",
		"1-nanoserver-sac2016",
		"1-nanoserver",
		"1-onbuild",
	}

	// TODO(Davy): Is there a way to decouple this from this test?
	err := testImage.getAuthToken()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	}

	err = testImage.getTags()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	}

	got := testImage.Tags
	fmt.Println("got:", got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %s | want: %s", got, want)
	}
}
