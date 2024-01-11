# Scrape the Docs

This repository contains the source code for a documentation scraping tool: Scrape the Docs!

Ever wanted to access web documentation for a software library or framework while you did not have access to the internet? This tool provides a convenient way to scrape web documentation and download it to your local machine so you can later use it when you are offline.

This tool crawls across a webpage starting from a specified root URL. Each page encountered is downloaded to the local machine. The HTML trees are transformed during execution to ensure that they will work on the local file system. JavaScript and CSS files are also downloaded (even if they are being served on an entirely different origin). The downloaded files can then be viewed through a browser as if you were online.

## Usage

To build the tool, run the following commands:

1. `git clone https://github.com/mattrrubino/scrape-the-docs.git`
2. `cd scrape-the-docs/main`
3. `go build -o scrape`
4. `./scrape {ROOT URL}`

Note that if you are on Windows, you must run `go build -o scrape.exe` in step 3 followed by `scrape.exe {ROOT URL}` in step 4. As an example, try the root url `docs.gl`. The documentation will be saved in `documentation/docs.gl`.
