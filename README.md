# Automaticly create Issues in Github from Google Cloud Logs 

Code of this repo is ment to be deploied to Google Cloud Function. 
It listen for PubSub message from Error Router "Sink" for `critical errors` (setup separatly)

![Screen Shot 2022-02-02 at 11 05 50 AM](https://user-images.githubusercontent.com/2000153/152116243-c022ea61-d5f1-4a92-b314-b7b2cea5f977.png)


___
## Requirements
Environment variables:
```
GITHUB_OWNER=name of user or organization that contains repository
GITHUB_TOKEN=Github token of user with privileges to create issues in the repository
GITHUB_SERVICES=```[
                  {
                    "serviceName": "your gcloud service name",
                    "repo": "your repository name for issues"
                  }, 
                  {
                    "serviceName": "your gcloud service name",
                    "repo": "your repository name for issues"
                  }
                ]```
 
ENV_TYPE=type of your environment (dev, prod, staging)
```
