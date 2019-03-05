@echo off
set ts3dir=%appdata%\TS3Client
set pluginname=voicybot-ts3plugin

if exist "%ts3dir%\plugins\%pluginname%.dll~" del "%ts3dir%\plugins\%pluginname%.dll~"
if exist "%ts3dir%\plugins\%pluginname%.dll" move "%ts3dir%\plugins\%pluginname%.dll" "%ts3dir%\plugins\%pluginname%.dll~"
if exist "%ts3dir%\plugins\%pluginname%.dll~" del "%ts3dir%\plugins\%pluginname%.dll~"
