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
- Shutdown channel is the global channel that shutsdown all the tab. The command channel is local channel that only has access to a particular tab

### User Input
- It is running in an individual Goroutine. (parent: main)

### Tabs
- New tab is like creating new Goroutine. (parent: main)
- 1 Tab can only serve 1 webpage at a time.
- The server that serves the webpage runs in a goroutine. (parent: tab)

### Webpage
- A webpage is simply a static HTML/CSS/JS directory.
- All the webpages are read from the `/webpages` directory.
- Two way to _remove_ or _add_ a new webpage to the browser:
	- You can paste the directory that contains the necessary files inside the `/webpages`.
	- Use the `Search` option and try searching for the particular webpage.

### Search
- It is running in an individual Goroutine. (parent: main)

### Synchronization
- It is running in an individual Goroutine. (parent: main)
- If another browser sends a message starting with "1" then that means they are checking if you are active and like to send the nodes copy you have
- If the message starts with "n" then it is the list of nodes, ip:port,ip:port...


## TO-DO

- Implement QUIC
- Clean variables like `Webpages` after user goes back to main user input.
- anyone can send nodes, without limit. add some rate limit to receiving nodes and the receivers.