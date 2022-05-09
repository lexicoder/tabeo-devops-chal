package spacex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrNotFound      error = errors.New("launchpad not found")
	ErrBadStatusCode error = errors.New("spacex returned invalid status code")
)

type Launch struct {
	LaunchPadID string `json:"launchpad"`
	Date        int64  `json:"date_unix"`
	// date_precision - Gives the date precision for partial dates.
	//Valid values are quarter, half, year, month, day, hour.
	DatePrecision string `json:"date_precision"`
}

func lastDayOfMonth(t time.Time) int {
	firstDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return lastDay.Day()
}

func (o *Launch) IsDayAvailable(t time.Time) (bool, error) {
	launchDate := time.Unix(o.Date, 0)
	start := time.Date(launchDate.Year(), launchDate.Month(), launchDate.Day(), 0, 0, 0, 0, time.UTC)
	var end time.Time
	switch o.DatePrecision {
	case "quarter":
		end = start.AddDate(0, 3, 0).AddDate(0, 0, -1)
	case "half":
		end = start.AddDate(0, 6, 0).AddDate(0, 0, -1)
	case "year":
		end = start.AddDate(1, 0, 0).AddDate(0, 0, -1)
	case "month":
		end = start.AddDate(0, 1, 0).AddDate(0, 0, -1)
	case "hour", "day":
		end = start
	}
	return t.After(end), nil
}

type LaunchPad struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

func (o *LaunchPad) IsActive() bool {
	return o.Status == "active"
}

type SpaceXClient struct {
	httpclient *http.Client
	baseUrl    string
}

func NewSpaceXClient(baseUrl string) *SpaceXClient {
	ans := SpaceXClient{
		httpclient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseUrl: baseUrl,
	}
	return &ans
}

func (o *SpaceXClient) IsLaunchpadAvailable(ctx context.Context, launchpadID string, ts time.Time) (bool, error) {
	launchpad, err := o.GetLaunchPadById(ctx, launchpadID)
	if err != nil {
		return false, err
	}
	if !launchpad.IsActive() {
		return false, nil
	}
	upcoming, err := o.QueryUpcomingLaunchesLaunchPad(ctx, launchpadID)
	if err != nil {
		return false, err
	}

	for i := range upcoming {
		available, err := upcoming[i].IsDayAvailable(ts)
		if err != nil {
			return false, err
		}
		if !available {
			return false, err
		}
	}
	return true, nil
}

func (o *SpaceXClient) GetLaunchPadById(ctx context.Context, launchpadID string) (LaunchPad, error) {
	var ans LaunchPad
	u := fmt.Sprintf("%s/%s/%s", o.baseUrl, "launchpads", launchpadID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ans, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := o.httpclient.Do(req)
	if err != nil {
		return ans, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return ans, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return ans, ErrBadStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ans, err
	}
	return ans, json.Unmarshal(body, &ans)
}

type launchesResp struct {
	Launches []Launch `json:"docs"`
}

// TODO follow pagination
// for the context of the task I just put a huge limit
// Since we are querying the upcoming launches of a specific launchpad we are probably covered for now
func (o *SpaceXClient) QueryUpcomingLaunchesLaunchPad(ctx context.Context, launchpadID string) ([]Launch, error) {
	u := fmt.Sprintf("%s/%s", o.baseUrl, "launches/query")
	q := o.buildUpcomingQuery(launchpadID)
	jsonBytes, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := o.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return nil, ErrBadStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var launchesRes launchesResp
	if err := json.Unmarshal(body, &launchesRes); err != nil {
		return nil, err
	}
	return launchesRes.Launches, nil
}

type searchQ struct {
	Query   map[string]interface{} `json:"query"`
	Options map[string]interface{} `json:"options"`
}

func (o *SpaceXClient) buildUpcomingQuery(launchpadID string) searchQ {
	searchQ := searchQ{
		Query:   make(map[string]interface{}),
		Options: make(map[string]interface{}),
	}
	searchQ.Options["sort"] = map[string]string{"date_unix": "asc"}
	searchQ.Options["select"] = []string{"launchpad", "date_unix", "date_precision"}
	searchQ.Options["limit"] = 10000
	searchQ.Query["upcoming"] = true
	searchQ.Query["launchpad"] = launchpadID
	return searchQ
}
