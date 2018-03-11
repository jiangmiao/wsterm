Websocket Terminal

    Run commands through websocket

Requirement

    golang - compile the code
    expect - handle processes
    bash   - run the command

Installation

    go get -u -v github.com/jiangmiao/wsterm/wsterm

Usage

    $ wsterm
    2018/03/11 19:39:28 websocket url ws://localhost:9300/ws

    # use https://github.com/hashrocket/ws send test requests
    # general echo
    $ ws ws://localhost:9300/ws
    > {"type":"exec", "data":"echo hello world"}
    < {"type":"stdout","data":"hello world\r\n"}
    < {"type":"exit","data":{"code":0,"message":""}}
    websocket: close 1000 (normal)

    # stderr handle
    $ ws ws://localhost:9300/ws
    > {"type":"exec","data":"curl"}
    < {"type":"stderr","data":"curl: try 'curl -"}
    < {"type":"stderr","data":"-help' or 'curl --manual' for more information\n"}
    < {"type":"exit","data":{"code":2,"message":"exit status 2"}}
    websocket: close 1000 (normal)

    # stop process
    $ ws ws://localhost:9300/ws
    > {"type":"exec", "data":"sudo ping 127.0.0.1"}
    < {"type":"stdout","data":"PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.\r\n64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.038 ms\r\n"}
    < {"type":"stdout","data":"64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.044 ms\r\n"}
    < {"type":"stdout","data":"64 bytes from 127.0.0.1: icmp_seq=3 ttl=64 time=0.043 ms\r\n"}
    # send stop message
    > {"type":"stop"}
    < {"type":"exit","data":{"code":130,"message":"exit status 130"}}
    websocket: close 1000 (normal)

Request Messages

    {"type": "exec", "data": "command"}
    {"type": "stop"}

Response Messages

    {"type": "stdout", "data": "stdout text"}
    {"type": "stderr", "data": "stderr text"}
    {"type": "exit",   "data": {"code": 0, "message", "error message"}}

License

    MIT
