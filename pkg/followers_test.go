package main

import (
	"fmt"
	"testing"
)

func TestFollower_String(t *testing.T) {
	type fields struct {
		Login          string
		Name           string
		DatabaseID     string
		FollowingCount float64
		FollowerCount  float64
		RepoCredit     float64
		Contributions  float64
		TotalCredit    float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test TestFollower_String",
			fields: fields{
				Login:          "LoginID",
				Name:           "",
				DatabaseID:     "MyID",
				FollowingCount: 5.0,
				FollowerCount:  10.0,
				RepoCredit:     100.0,
				Contributions:  30.0,
				TotalCredit:    20.0,
			},
			want: "{\n" +
				"\tLogin: " + "LoginID" + "\n" +
				"\tName: " + "" + "\n" +
				"\tDatabaseID: " + "MyID" + "\n" +
				"\tFollowingCount: " + fmt.Sprintf("%f", 5.0) + "\n" +
				"\tFollowerCount: " + "10.000000" + "\n" +
				"\tRepoCredit: " + "100.000000" + "\n" +
				"\tContributions: " + "30.000000" + "\n" +
				"\tTotalCredit: " + "20.000000" + "\n" +
				"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Follower{
				Login:          tt.fields.Login,
				Name:           tt.fields.Name,
				DatabaseID:     tt.fields.DatabaseID,
				FollowingCount: tt.fields.FollowingCount,
				FollowerCount:  tt.fields.FollowerCount,
				RepoCredit:     tt.fields.RepoCredit,
				Contributions:  tt.fields.Contributions,
				TotalCredit:    tt.fields.TotalCredit,
			}
			if got := f.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
