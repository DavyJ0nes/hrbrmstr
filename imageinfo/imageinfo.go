package imageinfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Version describes the version info of a Docker Image
type Version struct {
	Name    string
	ID      string
	Created time.Time
}

// byTime is used for custom sorting by sort.Sort
type byTime []Version

// RequestStat holds information about API Requests
type RequestStat struct {
	State   string
	Latency time.Duration
	URL     string
}

// ImageInfo describes a Docker Image
type ImageInfo struct {
	Repo         string
	baseURL      string
	startTime    time.Time
	Tags         []string
	Versions     []Version
	AuthToken    string
	RequestStats []RequestStat
}

// NewImageInfo inits a new Struct with relevant data from APIs
func NewImageInfo(repo string) (*ImageInfo, error) {
	i := &ImageInfo{
		Repo:      repo,
		baseURL:   "https://registry.hub.docker.com",
		startTime: time.Now(),
	}

	err := i.getAuthToken()
	if err != nil {
		return i, errors.Wrap(err, "Problems Getting Auth Token")
	}

	err = i.getTags()
	if err != nil {
		return i, errors.Wrap(err, "Issue Getting Tags for "+i.Repo)
	}

	for _, tag := range i.Tags {
		err = i.populateTagInfo(tag)
		if err != nil {
			return i, errors.Wrap(err, "Issue Populating Tag Info")
		}
	}

	return i, nil
}

// getTags calls docker hub to get all the tags associated with an image
func (i *ImageInfo) getTags() error {
	stat := RequestStat{}

	requestURL := fmt.Sprintf("%s/v2/%s/tags/list", i.baseURL, i.Repo)

	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return errors.Wrap(err, "Problem creating New GET Request")
	}
	req.Header.Add("Authorization", "Bearer "+i.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		stat.State = resp.Status
		stat.URL = requestURL
		stat.Latency = time.Since(i.startTime)
		i.RequestStats = append(i.RequestStats, stat)
		return errors.Wrap(err, "Problem Making API Request")
	}

	if resp.StatusCode != http.StatusOK {
		stat.State = resp.Status
		stat.URL = requestURL
		stat.Latency = time.Since(i.startTime)
		i.RequestStats = append(i.RequestStats, stat)
		return errors.New("Could Not Find Image")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Problem Reading Response")
	}
	defer resp.Body.Close()

	jsonStruct := struct {
		Tags []string `json:"tags"`
	}{}

	err = json.Unmarshal(body, &jsonStruct)
	if err != nil {
		return errors.Wrap(err, "Problem Decoding JSON")
	}

	// get first 10 elements
	i.Tags = jsonStruct.Tags[:9]

	stat.State = resp.Status
	stat.URL = requestURL
	stat.Latency = time.Since(i.startTime)
	i.RequestStats = append(i.RequestStats, stat)
	return nil
}

// getAuthToken calls the docker.io auth service to get an API auth Token
func (i *ImageInfo) getAuthToken() error {
	stat := RequestStat{}
	jsonStruct := struct {
		Token string `json:"token"`
	}{}

	requestURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", i.Repo)

	resp, err := http.Get(requestURL)
	if err != nil {
		stat.State = resp.Status
		stat.URL = requestURL
		stat.Latency = time.Since(i.startTime)
		i.RequestStats = append(i.RequestStats, stat)
		return errors.Wrap(err, "Problem making GET Request")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Problem parsing Response")
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, &jsonStruct)
	if err != nil {
		return errors.Wrap(err, "Problem Decoding JSON response")
	}

	i.AuthToken = jsonStruct.Token

	stat.State = resp.Status
	stat.URL = requestURL
	stat.Latency = time.Since(i.startTime)
	i.RequestStats = append(i.RequestStats, stat)
	return nil
}

func (i *ImageInfo) populateTagInfo(tag string) error {
	stat := RequestStat{}
	newVersion := Version{}
	timeInputFormat := "2006-01-02T15:04:05.999999999Z"
	requestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", i.baseURL, i.Repo, tag)
	client := &http.Client{}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return errors.Wrap(err, "Problem creating New GET Request")
	}
	req.Header.Add("Authorization", "Bearer "+i.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		stat.State = resp.Status
		stat.URL = requestURL
		stat.Latency = time.Since(i.startTime)
		i.RequestStats = append(i.RequestStats, stat)
		return errors.Wrap(err, "Problem Making API Request")
	}

	if resp.StatusCode != http.StatusOK {
		// return errors.New("Could Not Find Tag " + tag)
		// This needs to be handled better
		stat.State = resp.Status
		stat.URL = requestURL
		stat.Latency = time.Since(i.startTime)
		i.RequestStats = append(i.RequestStats, stat)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Problem parsing Response")
	}
	defer resp.Body.Close()

	jsonStruct := struct {
		Name    string `json:"name"`
		Tag     string `json:"tag"`
		History []struct {
			V1 string `json:"v1Compatibility"`
		} `json:"history"`
	}{}

	err = json.Unmarshal(body, &jsonStruct)
	if err != nil {
		return errors.Wrap(err, "Problem Decoding main JSON response")
	}

	versionStruct := struct {
		ID      string `json:"id"`
		Created string `json:"created"`
	}{}

	err = json.Unmarshal([]byte(jsonStruct.History[0].V1), &versionStruct)
	if err != nil {
		return errors.Wrap(err, "Problem Decoding verson JSON response")
	}

	// Set up Output
	newVersion.Name = fmt.Sprintf("%s:%s", jsonStruct.Name, jsonStruct.Tag)
	newVersion.Created, _ = time.Parse(timeInputFormat, versionStruct.Created)
	newVersion.ID = versionStruct.ID

	// Add to structs versions slice
	i.Versions = append(i.Versions, newVersion)

	stat.State = resp.Status
	stat.URL = requestURL
	stat.Latency = time.Since(i.startTime)
	i.RequestStats = append(i.RequestStats, stat)
	return nil
}

// String adheres to the Stringer interface so is able to be used by fmt
// to output in a meaningful format
func (i *ImageInfo) String() string {
	var outputStringSlice []string

	// sort Versions in ascending order by time Created
	sort.Sort(byTime(i.Versions))

	repoUnderline := strings.Repeat("-", len(i.Repo))
	outputStringSlice = append(outputStringSlice, fmt.Sprintf("hrbrmstr\n--------\n\nREPO: %s\n%s\n\nVersions\n--------", i.Repo, repoUnderline))

	// Output Version Info for Latest 10 Items
	for _, version := range i.Versions {
		// Use easier to read time format
		timeOutputFormat := version.Created.Format("02-01-2006 15:04:05")
		outputStringSlice = append(outputStringSlice, fmt.Sprintf("Name:\t %s\nCreated: %s\nID:\t %s\n", version.Name, timeOutputFormat, version.ID))
	}

	outputStringSlice = append(outputStringSlice, i.generateRequestStats())

	// bit of a hack but allows for easier output
	infoString := strings.Join(outputStringSlice, "\n")
	return infoString
}

func (i *ImageInfo) generateRequestStats() string {
	var (
		totalLatency  time.Duration
		totalRequests float64
		outputString  []string
	)

	stateMap := make(map[string]int)

	outputString = append(outputString, fmt.Sprint("\nRequest Stats\n-------------"))

	for _, stat := range i.RequestStats {
		stateMap[stat.State]++
		totalRequests++
		totalLatency = totalLatency + stat.Latency

	}

	for key, val := range stateMap {
		outputString = append(outputString, fmt.Sprintf("%s:\t %d", strings.ToUpper(key), val))
	}
	avgLatency := totalLatency.Seconds() / totalRequests
	outputString = append(outputString, fmt.Sprintf("\nTOTAL REQUESTS:\t %.0f\nAVG LATENCY:\t %.2fs", totalRequests, avgLatency))

	return strings.Join(outputString, "\n")
}

// Next 3 Functions are used to do Customer Sorting
// See here for interface explanation: https://golang.org/src/sort/sort.go?s=5414:5439#L14

// Len is used by sort.Sort for custom sorting
func (t byTime) Len() int { return len(t) }

// Swap is used by sort.Sort for custom sorting
func (t byTime) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// Less is used by sort.Sort for custom sorting
func (t byTime) Less(i, j int) bool { return t[i].Created.Before(t[j].Created) }
