const http = require('http')

let count = 100000
let thread = 100

let testTotal = 0
let errorCount = 0
let finished = 0

const options = {
    hostname: '127.0.0.1',
    port: 8080,
    method: 'GET'
}

http.request({ ...options, path: `/add/${count}` }).end()

function del() {
    const req = http.request({ ...options, path: `/del` }, res => {
        res.on('data', d => {
            const result = parseInt(d)
            if (result == 1) {
                testTotal++
                del()
            } else {
                finished++
                checkFinished()
            }
        })
    })

    req.on('error', e => {
        errorCount++
        checkFinished()
    })

    req.end()
}

function checkFinished() {
    if (finished + errorCount >= thread) {
        console.log(testTotal, finished, errorCount)
    }
}

function clientRun() {
    testTotal = 0
    errorCount = 0
    finished = 0

    for (let i = 0; i < thread; i++) {
        del()
    }
}

clientRun()
