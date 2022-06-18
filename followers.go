package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"text/template"
)

//go:embed templates
var fs embed.FS

type QL struct {
	Query string `json:"query"`
}
type Follower struct {
	Login          string
	Name           string
	DatabaseID     string
	FollowingCount float64
	FollowerCount  float64
	RepoCredit     float64
	Contributions  float64
	TotalCredit    float64
}

type FollowerArr []Follower

func (f FollowerArr) Len() int {
	//TODO implement me
	return len(f)
}

func isNaN(f float64) bool {
	return f != f
}
func (f FollowerArr) Less(i, j int) bool {
	//TODO implement me
	return f[i].TotalCredit < f[j].TotalCredit || (isNaN(f[i].TotalCredit) && !isNaN(f[j].TotalCredit))
}

func (f FollowerArr) Swap(i, j int) {
	//TODO implement me
	f[i], f[j] = f[j], f[i]
}

func (f Follower) String() string {
	return "{\n" +
		"\tLogin: " + f.Login + "\n" +
		"\tName: " + f.Name + "\n" +
		"\tDatabaseID: " + f.DatabaseID + "\n" +
		"\tFollowingCount: " + fmt.Sprintf("%f", f.FollowingCount) + "\n" +
		"\tFollowerCount: " + fmt.Sprintf("%f", f.FollowerCount) + "\n" +
		"\tRepoCredit: " + fmt.Sprintf("%f", f.RepoCredit) + "\n" +
		"\tContributions: " + fmt.Sprintf("%f", f.Contributions) + "\n" +
		"\tTotalCredit: " + fmt.Sprintf("%f", f.TotalCredit) + "\n" +
		"}"
}

func (f *Follower) setTotalCredit() {
	f.TotalCredit = f.FollowerCount + f.FollowingCount + f.RepoCredit + f.Contributions
}

func calculateRepoCredit(repos gjson.Result) float64 {
	var totalCredits float64
	for _, repo := range repos.Array() {
		forkCount := repo.Get("forkCount").Float()
		stargazerCount := repo.Get("stargazerCount").Float()
		totalCredits += forkCount
		totalCredits += stargazerCount
	}
	return totalCredits
}

func main() {
	var login string
	var pat string
	flag.StringVar(&login, "u", "", "GitHub ID")
	flag.StringVar(&pat, "p", "", "Personal Access Token")
	flag.Parse()

	fArr := FollowerArr{}

	client := &http.Client{}
	url := "https://api.github.com/graphql"

	tpl, err := template.ParseFS(fs, "templates/followers.ql")
	if err != nil {
		log.Fatal(err)
	}

	var hasNextPage bool
	after := ""
	for {
		var b bytes.Buffer
		err = tpl.Execute(&b, map[string]interface{}{
			"Login": login,
			"After": after,
		})
		if err != nil {
			log.Fatal(err)
		}
		ql := QL{
			Query: b.String(),
		}
		byteArr, err := json.Marshal(&ql)
		//fmt.Println(string(byteArr))
		if err != nil {
			log.Fatal(err)
		}
		reader := bytes.NewReader(byteArr)
		req, err := http.NewRequest("POST", url, reader)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Authorization", "Bearer "+pat)

		rep, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		data, err := ioutil.ReadAll(rep.Body)
		rep.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		ret := gjson.GetBytes(data, "data.user.followers")
		pageInfo := ret.Get("pageInfo")
		followers := ret.Get("nodes")
		for _, f := range followers.Array() {
			follower := Follower{
				Login:          f.Get("login").String(),
				Name:           f.Get("name").String(),
				DatabaseID:     f.Get("databaseId").String(),
				FollowingCount: f.Get("following.totalCount").Float() * 0.1,
				FollowerCount:  f.Get("followers.totalCount").Float() * 0.35,
				Contributions:  f.Get("contributionsCollection.contributionCalendar.totalContributions").Float() * 0.2,
				RepoCredit:     calculateRepoCredit(f.Get("repositories.nodes")) * 0.35,
			}
			follower.setTotalCredit()
			fArr = append(fArr, follower)
			//fmt.Println(len(fArr))
		}
		//fmt.Println(len(fArr))

		hasNextPage = pageInfo.Get("hasNextPage").Bool()
		after = "after: \"" + pageInfo.Get("endCursor").String() + "\""

		//fmt.Println(ret.Get("pageInfo.hasNextPage"))
		if !hasNextPage {
			break
		}
	}

	sort.Sort(sort.Reverse(fArr))
	for _, ff := range fArr {
		fmt.Println(ff)
	}

}
