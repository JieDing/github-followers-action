package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
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
func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func main() {
	var login string
	var pat string
	var followersConut int64
	flag.StringVar(&login, "u", "", "GitHub ID")
	flag.StringVar(&pat, "p", "", "Personal Access Token")
	flag.Parse()

	fArr := FollowerArr{}

	client := &http.Client{}
	url := "https://api.github.com/graphql"

	tpl, err := template.ParseFS(fs, "templates/followers.ql")
	htmlTPL, err := template.ParseFS(fs, "templates/table.tpl")
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
		err = rep.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		ret := gjson.GetBytes(data, "data.user.followers")
		pageInfo := ret.Get("pageInfo")
		followers := ret.Get("nodes")
		followersConut = ret.Get("totalCount").Int()
		for _, f := range followers.Array() {
			follower := Follower{
				Login:          f.Get("login").String(),
				Name:           f.Get("name").String(),
				DatabaseID:     f.Get("databaseId").String(),
				FollowingCount: f.Get("following.totalCount").Float() * 0.1,
				FollowerCount:  f.Get("followers.totalCount").Float() * 0.3,
				Contributions:  f.Get("contributionsCollection.contributionCalendar.totalContributions").Float() * 0.3,
				RepoCredit:     calculateRepoCredit(f.Get("repositories.nodes")) * 0.3,
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

	var rangeCount = min(18, followersConut)
	sTemp := strconv.FormatInt(rangeCount, 10)
	rangeInt, _ := strconv.Atoi(sTemp)
	html := "<table>\n"
	for i := 0; i < rangeInt; i++ {
		//fmt.Println(fArr[i])
		if i%2 == 0 {
			if i != 0 {
				html += "  </tr>\n"
			}
			html += "  <tr>\n"
		}
		var bf bytes.Buffer
		var name string
		if fArr[i].Name != "" {
			name = fArr[i].Name
		} else {
			name = fArr[i].Login
		}

		err = htmlTPL.Execute(&bf, map[string]interface{}{
			"login": fArr[i].Login,
			"id":    fArr[i].DatabaseID,
			"name":  name,
		})
		if err != nil {
			log.Fatal(err)
		}
		html += bf.String()
		//fmt.Println(bf.String())
	}
	html += "  </tr>\n</table>"
	//fmt.Println(html)
	str := "aaa<!--START_SECTION:top-followers-->hhh<!--END_SECTION:top-followers-->aaa"
	reg, err := regexp2.Compile("(?<=<!--START_SECTION:top-followers-->)[\\s\\S]*(?=<!--END_SECTION:top-followers-->)", 0)
	if err != nil {
		log.Fatal(err)
	}

	str, err = reg.Replace(str, "\n"+html+"\n", 10, 1)
	fmt.Println(str)
}
