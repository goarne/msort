consul agent -ui -dev 
msortweb --config-file=msortweb.d/appconfig.json
msortclient --config-file=./msortclient.d/appconfig.yaml

dig @127.0.0.1 -p 8600 msortweb.service.consul SRV
curl http://localhost:8500/v1/catalog/service/msortweb

docker rm msortweb
docker rmi goarne/msortweb

docker build -t goarne/msortweb .
docker run -d -p 8081:8081 --name msortweb goarne/msortweb


#docker with shell
docker run -i -t --name msortweb goarne/msortweb sh 


#dtart docker consul service
docker run -d  -p 8300-8302:8300-8302 -p 8400:8400 -p 8301-8302:8301-8302/udp -p 8600:8600 -p 8600:8600/udp -p 8500:8500 --name dev-consul consul

#Start registrator service which registers all docker containers
docker run -d --name=registrator --net=host --volume=/var/run/docker.sock:/tmp/docker.sock  gliderlabs/registrator:latest consul://localhost:8500

