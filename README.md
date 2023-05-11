# slackdump

Dumps info from slack.

Flags:

* -d='(value of your d cookie)
* -d-s='(value of your d-s cookie)'
* -token='(value of the results of the browser console command to run listed below)'


Console command:

```javascript
var localConfig = JSON.parse(localStorage.localConfig_v2)
localConfig.teams[localConfig.lastActiveTeamId].token
```
