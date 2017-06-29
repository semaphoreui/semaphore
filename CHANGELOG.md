
v2.4.0 / 2017-06-29
==============

  * update changelog, bump version to 2.4.0
  * Merge branch 'master' into develop
  * Merge pull request #370 from hsluoyz/patch-1
  * Update CONTRIBUTING.md with a note for Windows.
  * Merge pull request #371 from aioue/patch-1
  * Merge pull request #374 from strangeman/372-wrong-dates
  * fix wrong data format in Project activity log
  * Update main.go
  * Merge pull request #364 from KBraham/develop
  * Typo fix main.go
  * Merge pull request #359 from ansible-semaphore/feature/fix-login
  * rewrite login functions
  * update contributing.md
  * fix for base paths
  * base path resources
  * run migrations on startup
  * Merge pull request #355 from TeliaSweden/master
  * Merge pull request #345 from strangeman/alert-setting-343
  * Merge pull request #357 from ecornely/master
  * Merge pull request #342 from morph027/324-docker-zombies
  * Get tasks details
  * Fix nil pointer dereference when updating Template
  * add option for per-project telegram alert to different chats
  * fixes #324
  * Merge pull request #336 from strangeman/fix-auth-335
  * fix login logic when ldap is enabled
  * Merge pull request #330 from strangeman/fix-alerts-329
  * fix alert templates after 5bcb34e
  * set hostnaem
  * fix docker answer file
  * Update Dockerfile

v2.3.0 / 2017-04-19
===================

  * update changelog, bump version to 2.3.0
  * fix #323
  * add roadmap to readme
  * fix placeholder
  * fix #303
  * fix #312
  * fix tests
  * add octocat body
  * fix 2.2.1 migration #299
  * add default values for 2.3.0.sql
  * fixes for #310
  * fixes for #297
  * Merge branch 'develop' of github.com:ansible-semaphore/semaphore into develop
  * Merge pull request #299 from galexrt/improve-sql-error
  * improvements for #287
  * improve codebase after #275
  * fixes resulting from master merge
  * Merge branch 'master' into develop
  * Merge pull request #321 from serkin/master
  * Merge pull request #304 from z010107/master
  * Merge pull request #310 from strangeman/ldap-auth
  * Fixes #320
  * make ldap searched parameters configurable
  * merge with actual master
  * Merge pull request #307 from strangeman/telegram-alerts
  * add logrus logging, disable LDAP username and password editing on backend
  * make go vet happy
  * mispell
  * make username and password fields read-only for ldap users
  * add ldap settings to the setup process
  * add simple LDAP authentification to the config and login page
  * Merge pull request #298 from laeshiny/260
  * add response code check for telegram
  * add config generation for telegram alerting
  * add basic telegram alerting
  * Merge pull request #305 from strangeman/activity-log
  * add more verbosity about tasks to the Events description
  * Add extra validation for environment JSON
  * Add JSON validation in environment model
  * Fix the primary key creation queries Add id column to task__output table "instead" Print error message in case of database errors
  * Add sort, order parameter to Get Request of /project/id/(templates, inventory, environment, keys, repositories, users to api document
  * when requesting Team,  sort in ascending order by Name
  * Add sort, order parameter to Get Request of "project/id/users"
  * when requesting Playbook Repositories,  sort in ascending order by Name
  * Add sort, order parameter to Get Request of "project/id/repositories"
  * Removed comment
  * when requesting Key Store,  sort in ascending order by Name
  * Add sort, order parameter to Get Request of "project/id/key"
  * when requesting Environment,  sort in ascending order by Name
  * Add sort, order parameter to Get Request of "project/id/environment"
  * when requesting Task Templates,  sort in ascending order by Name
  * Add sort, order parameter to Get Request of "project/id/inventory"
  * Add missing prefix pt to query
  * when requesting Task Templates,  sort in ascending order by Alias
  * Add sort, order parameter to Get Request of "project/id/templates"
  * Merge branch 'develop' of github.com:ansible-semaphore/semaphore into 260
  * Merge pull request #297 from laeshiny/develop
  * correct to reformat from spaces to tabs
  * Merge pull request #300 from galexrt/improved-docker-entrypoint
  * Improve the docker entrypoint and dockerfile
  * Rearrange list of Task Template, Inventory, Environment, Team in UI
  * Add css (margin-left: 5px) to button between copy and run
  * Add page title to class at ui-view
  * Merge pull request #287 from strangeman/email-alerts
  * Merge remote-tracking branch 'origin/259' into develop
  * update swagger docs with models changes
  * made changes from review
  * Merge pull request #286 from strangeman/empty-cli-args
  * Merge pull request #292 from laeshiny/develop
  * Merge pull request #294 from galexrt/sql-primary-keys
  * Added v2.2.1 sql migration file that adds primary keys to the tables
  * Add copy feature to task templates.
  * Add copy button at Task Templates page
  * Fix duplicated mapping key mapping key "Task"
  * Merge pull request #289 from laeshiny/develop
  * english muthafucka do you speak it!?
  * correct the response code in api document
  * correct the response content and code
  * add content to response of post /project/{project_id}/inventory
  * add .idea/ to .gitignore for Pycharm
  * fix user alerts updating
  * add new config parameters to the setup procedure
  * Merge branch 'master' into email-alerts
  * add alert setting for project
  * provide NULL instead of empty string, when Extra CLI Arguments was deleted
  * add alert setting for user and (WIP) project
  * Merge branch 'master' into develop
  * Merge pull request #283 from laeshiny/master
  * add missing package and command
  * use html/template for mail subject and body
  * move mail sending logic to util package
  * Merge branch 'master' into develop
  * Merge pull request #277 from strangeman/wrong-time
  * Use UTC_TIMESTAMP instead of NOW
  * Merge pull request #275 from commodityvectors/sshcert
  * Added SSH certificate support
  * merge models -> db
  * 🎉 gin -> net/http
  * moar refactor
  * begin refactor gin -> net/http
  * update Dockerfile, changelog & release scripts
  * [WIP] add alerts for failed deploy

v2.2.0 / 2017-02-22
===================

  * bump version to 2.2.0
  * update add templates, gh release script
  * compile with go1.8
  * Merge branch 'master' of github.com:ansible-semaphore/semaphore
  * update contributing.md
  * Merge pull request #257 from strangeman/dashboard-alias
  * Merge pull request #262 from strangeman/templateid-ui
  * Merge pull request #268 from nightvisi0n/fix_go-github-api-break
  * fix api breaking of google/go-github
  * Merge pull request #265 from ansible-semaphore/fix-264
  * Reload on modal dismiss
  * Merge pull request #258 from strangeman/docs-improve
  * Add/Update Template dialog: Add Template ID field, mark some fields as required
  * Add task template name to log too
  * Small documentation improvements
  * Add task template names for dashboard
  * Merge pull request #256 from kpashka/master
  * Remove trailing dot-slash in find output
  * Copy-paste fixes
  * It's method, not a function
  * Use temp path for update repository function
  * Pass OS environment variables to Ansible
  * Added link to discord
  * Merge pull request #244 from jerrygb/patch-1
  * Update CONTRIBUTING.md
  * Merge pull request #241 from pianzide1117/master
  * fix error route   /project/{project_id}/template   ==>   /project/{project_id}/templates

## v2.1.0 | 22-12-2016

  * fix #202
  * update api docs
  * fix #214 - events api
  * parse time with momentjs - fix #197
  * Merged branch master into master
  * fix circle & docker hub
  * Merge pull request #231 from ringtail/master
  * Merge pull request #157 from tokuhirom/galaxy
  * remove chat room
  * Merged branch master into master
  * fix #183
  * fix #193 - auth middleware bug
  * Merge pull request #235 from rakshazi/patch-1
  * Task history: changed status colors
  * Update CONTRIBUTING.md
  * add chat room link to readme
  * prevent removing last admin from project
  * fix migration file
  * attempt to fix circle CI
  * improve runner code
  * Merge remote-tracking branch 'refs/remotes/knsr/repo-tags'
  * Merge pull request #228 from gcavalcante8808/master
  * Alias field was not created on the DB. Corrected: added version 1.9.0.
  * Alias field was not created on the DB. Corrected: added version 1.9.0.
  * Merge pull request #211 from Woorank/sshclient
  * Merge pull request #207 from jahantech/master
  * Merge pull request #224 from gcavalcante8808/issue_188
  * Fixes #188.
  * Corrected display information of step 4.
  * Fixes #213. A set of information about development added to guide.
  * Add ssh client to dockerfile
  * Erorr -> Error
  * Merge pull request #205 from Woorank/docker
  * Updated compose file to version 2, moved variables from dockerfile to compose file, pointed towards future automated build
  * Moved dockerfile to root and changed base image to Alpine
  * correct import
  * add support for git repo tags or branches
  * Merge pull request #181 from modoojunko/master
  * add missing space
  * Update runner.go
  * Merge pull request #178 from goozbach/dockerfiles
  * Adding readme
  * fixing origin header break and websockets proxy issues
  * docker compose working
  * clean up packages
  * mostly done, just no internet
  * Merge pull request #160 from tokuhirom/clear-repository-cache
  * Merge pull request #156 from tokuhirom/dry_run
  * Merge pull request #155 from tokuhirom/show-repository-info
  * Clear repository cache after update/delete repository information. Close #159
  * Added Ansible Galaxy support. Close #150
  * Added dry_run button. close #152
  * Show repository url in log

## v2.0.4 | 28-6-2016

- Show user name for tasks (thanks @tokuhirom!)
- Fix critical bug with creating things

## v2.0.3 | 25-6-2016

- Much better UI
- Editable Project
- Fix SQL bug
- Admin privileges in projects are more relevant

## v2.0.2 | 23-5-2016

- Improve upgrade process (fixes #106)
- Improve upgrade UI
- Delete Users API
- Fetch User API
- User update API & UI
- Security improvement (does not spill access key secret over api)
- Improve setup (fixes #100)
- Fix sql migrations for new setups

## v2.0.1 | 20-5-2016

- Add details to contribution guide
- Fix sql errors with project creation (resolves #92, #83)
- Check for updates every day
- Display alert next to settings cog when update is available
- Fixed bugs #91, #93

## v2.0 | 17-5-2016

- Removed redis dependency
- Upgrade semaphore from UI
- System info page
- new branding
- raw output option for logs (no time prefix)
- fix ws logging
- status updates from ws

## v2.0-beta-2 | 29-4-2016

- Fix SQL migrations to work under strict mode (#81)
- Minor UI improvements

## v2.0-beta | 26-4-2016

- Better `-setup`
- Testing auto-upgrades
