<div align="center">
  
# QuickURL

Sharing files super easy. Just give it the filepaths, and QuickURL generates URLs with all available entry points.

[![License](https://img.shields.io/github/license/fpfeng/quickurl)](https://opensource.org/licenses/Apache-2.0)
[![quickurl](https://github.com/fpfeng/quickurl/workflows/build/badge.svg)](https://github.com/fpfeng/quickurl/actions/workflows/build.yml)
[![downloads](https://img.shields.io/github/downloads/fpfeng/quickurl/total?color=green)](https://github.com/fpfeng/quickurl/releases)

https://github.com/fpfeng/quickurl/assets/2508212/b1a7f947-8e59-4345-a151-fe4b29571f7f

</div>

## Installation
Go to [Release](https://github.com/fpfeng/quickurl/releases) to download  compiled binary.
```bash
# for linux amd64
wget -O quickurl https://github.com/fpfeng/quickurl/releases/latest/download/quickurl_linux_amd64
sudo mv quickurl /usr/local/bin/quickurl && chmod +x /usr/local/bin/quickurl
```

## Get Started
```bash
# will serving those files on default port 5731
quickurl /path/to/file1 /path/to/file2

# custom port number
quickurl -s /path/to/file1 -s /path/to/folder1 -p 8080

# print public ips only
quickurl -s /path/to/file1 -s /path/to/file2 -public-ip
```

## Development Roadmap
- ~~Support folders~~
- ~~Support download items at once~~
- Support upload
- ~~Fetch public IPs from external API~~
- ~~Update by itself~~
- Testing
- Improve output such as colorful result, support JSON and YAML
- Install script
- Request log
