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
Usage:  /Users/tamareu/Code/git-stats/git-stats -repo=repo_path [options]
  -config string
    	[optionnal] Path to the configuration file
  -help
    	[optionnal] Displays this helps and quit
  -repo string
    	[mandatory] Path to the git repository
  -subtree string
    	[optionnal] Subtree you want to parse (default "/")
```

![Alt text](/screenshot.png?raw=true "Preview")
