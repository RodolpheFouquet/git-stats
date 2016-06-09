[![Build
Status](https://travis-ci.org/RodolpheFouquet/git-stats.svg?branch=master)](https://travis-ci.org/RodolpheFouquet/git-stats)

# Git-stats

## How to build:
To install dependencies install gpm https://github.com/pote/gpm
```
wget https://raw.githubusercontent.com/pote/gpm/v1.4.0/bin/gpm && chmod
+x gpm && sudo mv gpm /usr/local/bin
```

And type
```
gpm install 
```

WORK IN PROGRESS

Dumps the number of additions and deletions from a repostory or a
subtree of the repository

```
Usage:  git-stats repo_path subtree
Example:  git-stats repo_path  / will  give the stats for the whole repository
          git-stats repo_path /module1/src  will give the stats for the module1/src subpath
```

![Alt text](/screenshot.png?raw=true "Preview")
