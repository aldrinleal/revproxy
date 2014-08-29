FROM ubuntu:14.04

RUN sudo apt-get update ; sudo apt-get -y install curl unzip
RUN bash -xe -c "curl -vLo /tmp/revproxy.zip 'http://beta.gobuild.io/download?os=linux&arch=amd64&rid=329' ; mkdir /app ; cd /app ; unzip /tmp/revproxy.zip"

EXPOSE 80

CMD /app/revproxy -port 80
