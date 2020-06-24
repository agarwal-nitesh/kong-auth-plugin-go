### Go plugin for kong

This plugin is based on the new Go based interface supported by kong.
This plugin authorizes a request by intercepting it at the gateway level.
This allows any request to be authorized before routing it upstream by just enabling this plugin on the route/service.
For demo purposes, it also authorizes upto 3 roles which can be specified while enabling plugin at kong.

#### Script to setup go-pluginserver

Note: This script clones the pluginserver and builds as there is a bug that is currently not closed on go-pluginserver github.
```
#!/bin/bash
cd ~/ 
git clone https://github.com/Kong/go-pluginserver.git 
cd go-pluginserver 
/usr/local/go/bin/go get github.com/Kong/go-pdk@master 
/usr/local/go/bin/go build
sudo cp ~/go-pluginserver/go-pluginserver /usr/local/bin/
```

```
#!/bin/bash
sudo mkdir -p /usr/local/share/go/1.14/kong/plugins
sudo chown -R ubuntu /usr/local/share/go/1.14/kong
```
#### Deploy Auth Plugin
```
#!/bin/bash
export KONG_GO_PLUGIN_DIR=/usr/local/share/go/1.14/kong/plugins
cd ~/zkrull_auth_plugin && git pull origin master && sh .bin/build.sh
cp ~/zkrull_auth_plugin/zkrull-auth.so $KONG_GO_PLUGIN_DIR
kong stop
/usr/local/bin/kong start -c /home/ubuntu/zkrull_server_setup/kong/kong.conf 2>&1 > /home/ubuntu/logs/kong/kong.log
```

#### kong.conf file configuration:
```
plugins = bundled,zkrull-auth
go_plugins_dir = /usr/local/share/go/1.14/kong/plugins 

```
