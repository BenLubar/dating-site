FROM golang

COPY . /src
RUN cd /src \
 && go build -o /dating-site

ENV PATH=/
WORKDIR /
VOLUME /data
EXPOSE 80

CMD ["dating-site"]
