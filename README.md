# QuantumScraper ⚡️ 🚀

QuantumScraper is a blazing fast scraper specifically built to crawl the entire web at an extremely fast pace.
This is not your average, general-purpose scraper. QuantumScraper is focused on speed and efficiency, delivering web-scale crawls at lightning-fast speeds.

## Features 🌟

- Built-in **NSQD** server for URL queuing
- Advanced URL management and storage using **BadgerDB**
- Fast HTTP networking with **fasthttp**



## Data Preparation

Setup your AWS CLI, then download latest CommonCrawl Index

    aws s3 sync s3://commoncrawl/cc-index/table/cc-main/warc/crawl=CC-MAIN-2023-14/subset=warc/ cc-index/