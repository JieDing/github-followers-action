query {
  user(login: "{{ .Login }}") {
  	followers(first: 2,{{ .After}}){
      pageInfo{
        endCursor
        hasNextPage
      }
      nodes {
          login
          name
          databaseId
          following {
              totalCount
          }
          repositories(first: 3, isFork: false, orderBy: {
              field: STARGAZERS,
              direction: DESC
          }) {
              totalCount
              nodes {
                forkCount
                stargazerCount
              }
        	}
        	followers {
              totalCount
          }
          contributionsCollection{
              contributionCalendar{
                totalContributions
              }
          }
      }
      totalCount

    }
  }
}