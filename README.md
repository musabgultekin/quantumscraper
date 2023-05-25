# QuantumScraper ⚡️ 🚀

QuantumScraper is a blazing fast scraper specifically built to crawl the entire web at an extremely fast pace.
This is not your average, general-purpose scraper. QuantumScraper is focused on speed and efficiency, delivering web-scale crawls at lightning-fast speeds.

## Features 🌟

- Ultra-fast HTTP networking with **fasthttp**
- Built-in **NSQD** server for URL queuing
- Advanced URL management and storage using **BadgerDB**

# Setup

## Requirements

- Go
- AWS CLI (Optional)

## URL Starter Data Preparation

To start off the scraper, we're initializing the scraper's state with the URLs of the CommonCrawl.
While we don't need this, we're doing to speedup the scraper. We could start from the root domains priovided by [Domains Index](https://domains-index.com/) , [Zonefiles](https://zonefiles.io/), [Domains Monitor](https://domains-monitor.com/) or similar.

Setup your AWS CLI, login to your AWS account. 
Then download latest columnar CommonCrawl Index. (This might incur bandwidth costs)

    aws s3 sync s3://commoncrawl/cc-index/table/cc-main/warc/crawl=CC-MAIN-2023-14/subset=warc/ cc-index/

Without AWS CLI: Visit this link and download each of them one by one.

    https://us-east-1.console.aws.amazon.com/s3/buckets/commoncrawl?prefix=cc-index%2Ftable%2Fcc-main%2Fwarc%2Fcrawl%3DCC-MAIN-2023-14%2Fsubset%3Dwarc%2F&region=us-east-1

Without AWS CLI and AWS Account (Slow):

1. Visit: https://data.commoncrawl.org/crawl-data/index.html
2. Visit the latest crawl, for example: https://data.commoncrawl.org/crawl-data/CC-MAIN-2023-14/index.html
3. Download Columnar URL index files: For example: https://data.commoncrawl.org/crawl-data/CC-MAIN-2023-14/cc-index-table.paths.gz
4. Unzip with gzip, and append "https://data.commoncrawl.org/" to each line and download all the files.

### Manual Download


## Increase OS Limits

    ulimit -n 1048576


## Start scraping

    go run main.go