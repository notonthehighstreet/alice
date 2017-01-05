NAME=autoscaler
HUB=hub.noths.com

all: test localbuild dockerbuild

test:
	docker run --rm=true --volume="${CURDIR}:/go/src" golang:1.7-alpine /bin/sh -c \
	cd /go/src && \
	glide install && \
	go test ./manager/...

localbuild:
	docker run --rm=true --volume="${CURDIR}:/go/src" -w /go/src golang:1.7-alpine /bin/sh -c \
		cd /go/src && \
		go get && \
		glide install && \
		go build --ldflags '-s -w -linkmode external -extldflags "-static"' -o autoscaler main.go && \
		chmod 755 autoscaler

dockerbuild: localbuild
	docker build -t ${HUB}/${NAME}:latest .

dockerpush: dockerbuild
	docker push ${HUB}/${NAME}:latest
