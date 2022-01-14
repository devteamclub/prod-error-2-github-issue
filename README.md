# prod-error-2-github-issue

This repo contains code that can be deployed as Google Cloud Function. 
It receives cloud run critical errors via PubSub 
message and post an issue into Github repository.
___
## Requirements
Environment variables:

GITHUB_OWNER=name of user or organization that contains proper repository<br>
GITHUB_TOKEN=Github token of user with privileges to create issues in the repository <br>
