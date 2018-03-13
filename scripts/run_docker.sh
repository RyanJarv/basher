#!/usr/bin/env bash
docker run -it -v $(pwd):/go/src/github.com/RyanJarv/bashfix -w /go/src/github.com/RyanJarv/bashfix bashfix bash -l
