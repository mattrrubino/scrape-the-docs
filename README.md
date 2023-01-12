# Scrape the Docs

This repository contains the source code for a documentation scraping tool: Scrape the Docs!
This tool crawls across a webpage starting from a specified root URL. Each page encountered
is downloaded to the local machine. The HTML trees are transformed during execution to ensure
that they will work on the local file system. JavaScript and CSS files are also downloaded
(even if they are being served on an entirely different origin). This allows the scraped
documentation to be used offline, while creating the illusion of internet connectivity!
