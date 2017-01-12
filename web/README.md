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
