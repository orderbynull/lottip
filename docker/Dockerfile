FROM frolvlad/alpine-go

RUN apk update && apk upgrade && \
    apk add --no-cache bash git

RUN git clone http://github.com/rkoshy/lottip
WORKDIR lottip
RUN go get github.com/mjibson/esc
RUN go install github.com/mjibson/esc
RUN /root/go/bin/esc -o fs.go -prefix web -include=".*\.css|.*\.js|.*\.html|.*\.png" web
RUN go get
RUN go install
WORKDIR /
RUN rm -rf /root/go/bin/esc && rm -rf lottip

ADD run.sh /run.sh
RUN chmod +x /run.sh

EXPOSE 4041
EXPOSE 9999

CMD /run.sh
