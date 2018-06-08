# discovery

## a simple discovery server for miiverse clones and replacements

### instructions

- install [golang](https://golang.org) on your server

- run `go get -u gitlab.com/superwhiskers/discovery && cd ~ && mkdir discovery && cd discovery && cp $GOROOT/bin/discovery . && cp $GOROOT/src/gitlab.com/superwhiskers/discovery/config.example.yaml ./config.yaml` on your server

- edit the config.yaml file in the current folder to your liking, and place it behind a reverse proxy (set the line that says `https: true` to `https: false` if you are going to do this) if you are running more than one server on the same box

### support

dm `superwhiskers#3210` on discord for help