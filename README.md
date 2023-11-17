# GoSSLChecker
A quick-and-dirty Go program that checks for SSL/TLS certificates expiry dates and reports them to the specified Telegram chat.

## Current status
4/5 (mostly done/beta version). Buildable and usable, but there is certainly more room for improvement.

## Usage
```
   SSLChecker options url1 [url2 ...]

OPTIONS:                                                                                                             
   --min-days value, -m value    minimal remaining active days for a certificate (default: 5)
   --send-delay value, -d value  delay between message sending attempts (in seconds) (default: 3)
   --max-tries value, -x value   maximum number of message sending attempts (default: 5)
   --tgm-token value, -t value   Telegram token for sending messsages
   --tgm-chatid value, -c value  Telegram chat id for sending messsages
   --lang-en, -e                 force usage of English language, instead of cheking the OS defaults
   --lang-ru, -r                 force usage of Russian language, instead of cheking the OS defaults
   --verbose, -v                 verbose mode
   --help, -h                    print program usage information and exit
   --version, -V                 print program version and exit
```
