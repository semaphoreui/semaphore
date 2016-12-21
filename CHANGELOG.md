
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