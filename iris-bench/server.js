const http = require('http')

let total = 0

function server(req, resp) {
    if (req.url === '/add/1000') {
        total += 1000
    } else if (req.url === '/add/10000') {
        total += 10000
    } else if (req.url === '/add/100000') {
        total += 100000
    } else if (req.url === '/add/1000000') {
        total += 1000000
    } else if (req.url === '/add/10000000') {
        total += 10000000
    }
    if (req.url === '/del') {
        if (total > 0) {
            total--
            resp.end("1")
        } else {
            resp.end("0")
        }
    } else {
        resp.end("" + total)
    }
}

http.createServer(server).listen(8080, 'localhost')
