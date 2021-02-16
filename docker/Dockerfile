FROM frolvlad/alpine-go


RUN apk update && apk upgrade && \
    apk add --no-cache bash git

RUN go get github.com/brainz80/lottip && \
    go get github.com/mjibson/esc && \
    go install github.com/mjibson/esc && \
    cd /root/go/src/github.com/brainz80/lottip && \
    /root/go/bin/esc -o fs.go -prefix web -include=".*\.css|.*\.js|.*\.html|.*\.png" web && \
    go build

ADD run.sh /run.sh
RUN chmod +x /run.sh

EXPOSE 4041
EXPOSE 9999

CMD /run.sh



