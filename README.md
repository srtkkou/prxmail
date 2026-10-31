# prxmail

Very simple command-line smtp email sender.
Mainly used to proxy email via postfix alias.

## Install

1. Clone git repository.
```bash
$ git clone https://github.com/srtkkou/prxmail.git
```

2. Build `prxmail.exe` using [task](https://github.com/go-task/task) command.
```bash
$ task build
```

3. Copy `prxmail.exe` and `prxmail.toml` to user home direcotry.
```bash
$ cp prxmail.exe /home/user/
$ cp prxmail.toml.sample /home/user/prxmail.toml
```

4. Edit `prxmail.toml`.

5. Create `.forward` flie in user home directory.
```bash
$ touch /home/user/.forward
$ chmod 600 /home/user/.forward
```

6. Add following line on `.forward` file.
```
"| /home/user/prxmail.exe -s SUBJECT -r SENDER_ADDRESS -t RECEIVER_ADDRESS > /dev/null 2>&1"
```
Note: Sending `STDOUT` and `STDERR` to `/dev/null` to avoid
`postfix` from resending email on error.

7. Restart `postfix`. (Maybe unnecessary.)
  ```bash
$ sudo newaliases
$ sudo systemctl restart postfix
```
