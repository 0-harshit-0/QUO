# QUO

A CLI Browser


## USAGE

- To operate it type the command id/index.
- To search type the `6` command and press enter.
	- Now, search for the page that you are looking for and press "Enter".
	- To run the page, type `-y {id}` the id is the page id/index as shown in the CLI.
	- To close the search query command type `-n`.
- To close the browser type `0`.


## Concepts

### Browser
- 

### User Input
- It is running in an individual Goroutine. (parent: main)

### Tabs
- New tab is like creating new Goroutine. (parent: main)
- 1 Tab can only serve 1 webpage at a time.
- The server that serves the webpage runs in a goroutine. (parent: tab)

### Webpage
- A webpage is simply a static HTML/CSS/JS directory.
- All the webpages are read from the `/webpages` directory.
- Three way to _remove_ or _add_ a new webpage to the browser:
	- You can paste the directory that contains the necessary files inside the `/webpages`.
	- Use the `Sync Webpages` option after running the browser.
	- Use the `Search` option and try searching for the particular webpage.

### Search
- It is running in an individual Goroutine. (parent: main)

### Synchronization
- It is running in an individual Goroutine. (parent: main)


## TO-DO

- Clean variables like `Webpages` after user goes back to main user input.
- anyone can send nodes, without limit. add some rate limit to receiving nodes and the receivers.