consul agent -ui -dev 

msortweb --config-file=msortweb.d/appconfig.json

dig @127.0.0.1 -p 8600 msortweb.service.consul SRV


curl http://localhost:8500/v1/catalog/service/msortweb


 msortclient --config-file=./msortclient.d/appconfig.yaml  