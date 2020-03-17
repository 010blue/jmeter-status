# JMeter Status

## install dependencies
    # web
    cd web && npm install

    # golang, using dep
    cd service && dep ensure

## configure
    cd service
    mv config.example.toml config.toml
    vi config.toml # change to your own information
    # for mysql
    create database dbname
    execute config/db.sql to mysql database
    
## fetch build for linux
    cd service
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build fetch.go 

## fetch data
    cd service
    go build fetch.go
    # manual execution
    ./fetch

    # scheduled tasks, using crontab, every 10 minutes
    */10 * * * * cd /path/service && ./fetch

## start web...
use nginx or others to point to web directory

![Image text](https://raw.githubusercontent.com/010blue/jmeter-status/master/example/screenshot.png)