# Postfix Virtual Alias Manager
## Description
View, create, delete and search aliases in your postfix alias database

![Alt text](/images/browser.jpg?raw=true)

## Build

##### Windows (CMD)
```cmd
git clone https://github.com/guitarmarx/postfix_virtual_alias_manager.git
cd postfix_virtual_alias_manager
set GOOS=windows
set GOARCH=amd64
go build
```

##### Linux
```sh
git clone https://github.com/guitarmarx/postfix_virtual_alias_manager.git
cd postfix_virtual_alias_manager
GOOS=windows GOARCH=amd64 go build
```
## Configuration
Create `conf.json` in the directory where the executable is located
Content:

```sh
{
    "ServerPort": "8000",
    "DbHost": "<Host>",
    "DbPort": "<port>",
    "DbUser": "<user>",
    "DbPassword": "<password>",
    "DbName": "postfix",
    "AliasTableName": "virtual_aliases"
}
```