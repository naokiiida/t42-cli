```bash
go env
export GOPATH=$(go env)
REPO="github.com/naokiiida/t42-cli"
NEW_REPO_PATH="$GOPATH/src/$REPO"
mkdir -p $NEW_REPO_PATH
cd $NEW_REPO_PATH
go mod init $REPO
cobra-cli init

//test
go run main.go
go build .
CLI_NAME=$(basename $REPO)
./$CLI_NAME
```