# slackdump

Dumps info from slack.

```shell
slackdump

Usage:
  slackdump [flags]
  slackdump [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command
  list-conversations list-conversations

Flags:
  -d, --d-cookie string    'd' cookie value
  -s, --ds-cookie string   'd-s' cookie value
  -h, --help               help for slackdump
  -t, --token string       slack token (see readme)

Use "slackdump [command] --help" for more information about a command.
```

Flags Descriptions:

* `--d-cookie='(value of your d cookie)`
* `--ds-cookie='(value of your d-s cookie)'`
* `--token='(value of the results of the browser console command to run listed below)'`

Console command:

```javascript
var localConfig = JSON.parse(localStorage.localConfig_v2)
localConfig.teams[localConfig.lastActiveTeamId].token
```
