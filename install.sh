#!/bin/sh -e
ts3dir="$HOME/.ts3client"
pluginname="voicybot-ts3plugin"
mkdir -p "${ts3dir}/plugins"
go build -v -buildmode=c-shared -o "${ts3dir}/plugins/${pluginname}.so"
