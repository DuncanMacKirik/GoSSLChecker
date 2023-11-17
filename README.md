# GoSSLChecker
A quick-and-dirty Go program that checks for SSL/TLS certificates validity and expiry dates and reports them to the specified Telegram chat.

## Current status
4/5 (mostly done/beta version). Buildable and usable, but there is certainly more room for improvement.

## Some history
It's my first Go program, written with the following goals in mind:
1) test the language, its runtime and external libraries in real-world scenario, including building cross-platform applications
2) replace similar, but not-so-reliable Python and PHP solutions
3) see how hard it is to quickly implement simple internationalization and (at least partial) POSIX compliance in Go
4) practice using Git and Github some more
IMHO all these goals have been accomplished, with maybe some minor improvements left to implement when there's enough time.

## Usage
<pre>
<b>SSLChecker</b> options url1 [url2 ...]

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
</pre>
  
# GoSSLChecker
Программа с открытым кодом на Go, проверяющая корректность и сроки истечения, и рапортующая о найденных проблемах в Telegram.

## Состояние проекта
4/5 (в основном готово/бета-версия). Можно собирать и пользоваться, но ещё есть что улучшать.

## Зачем это вообще?
Это моя первая программа на Go, и цели перед её разработкой ставились следующие:
1) испытать в деле сам язык, его инструментарий и библиотеки, причём желательно сразу для нужной в реальной жизни задачи;
2) заменить предыдущие (и менее надёжные) решения на интерпретируемых языках - PHP и Python;
3) изучить, насколько трудно в минимальные сроки добавить в подобный проект многоязычность и (хотя бы частичную) совместимость с POSIX;
4) дополнительно потренироваться в использовании Git и Github.
С моей точки зрения, все эти цели были успешно достигнуты. Остались мелкие улучшения, которые планируется вносить по возможности.

## Параметры
<pre>
<b>SSLChecker</b> options url1 [url2 ...]

ОПЦИИ:
   --min-days value, -m value    минимально допустимое время истечения сертификата (в днях) (default: 5)
   --send-delay value, -d value  длительность задержки между попытками отправки сообщений в Telegram (в секундах) (default: 3)
   --max-tries value, -x value   максимальное количество попыток отправки сообщений в Telegram (default: 5)
   --tgm-token value, -t value   значение токена Telegram для отправки сообщений
   --tgm-chatid value, -c value  значение chat id Telegram для отправки сообщений
   --lang-en, -e                 использовать английский язык (вместо попытки автоопределения языка ОС)
   --lang-ru, -r                 использовать русский язык (вместо попытки автоопределения языка ОС)
   --verbose, -v                 включить вывод подробной информации
   --help, -h                    показать короткую справку об использовании программы и выйти
   --version, -V                 показать версию программы и выйти

ВЕРСИЯ:
   0.5.0
</pre>
