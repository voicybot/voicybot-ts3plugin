@echo off
set ts3dir=%appdata%\TS3Client
set pluginname=voicybot-ts3plugin

go build -v -buildmode=c-shared -o "%ts3dir%\plugins\%pluginname%.dll"
